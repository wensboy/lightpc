// Package async 提供一组轻量, 显式的异步原语(Future / Scope / 组合器).
//
// 注意: 若业务 fn 无视 ctx 而永久阻塞, 本包无法回收该 goroutine-
// 这是 Go 的固有约束; 本包保证的是"自身永不泄漏", 并把 ctx 显式交给 fn.
package async

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"
)

// Future 表示一个将来会产生 (T, error) 的异步任务句柄.
//
// 所有方法均可并发调用.Get 系列方法可被多次, 多 goroutine 调用,
// 均返回同一结果; 完成后再次 Get 立即返回.
type Future[T any] interface {
	// Get 阻塞直到任务完成, 返回结果或错误.
	// 错误来源包括: fn 返回的 error, fn panic(已转为 error),
	// 父 ctx 被取消(若 fn 返回了 ctx.Err()).
	Get() (T, error)

	// GetContext 带取消/超时地等待结果.
	// 若 ctx 先于任务完成而结束, 立即返回 ctx.Err();
	// 后台任务不受影响, 仍会继续执行至自然结束.
	GetContext(ctx context.Context) (T, error)

	// TryGet 非阻塞探测.ok=false 表示任务尚未完成.
	TryGet() (val T, err error, ok bool)

	// Done 返回一个在任务完成时被关闭的只读 channel,
	// 便于与 select 组合(超时, 多路复用等).
	Done() <-chan struct{}

	// Cancel 请求取消任务.任务是否尽快退出取决于 fn 是否响应 ctx.
	// 幂等.
	Cancel()
}

// Scope 是 errgroup 风格的生命周期容器.
//
// 语义: 任一子任务返回非 nil 错误时, Context() 被取消,
// 其余子任务应感知退出; Wait 返回第一个非 nil 错误.
type Scope interface {
	// Context 返回作用域上下文, 供子任务使用.
	// 首次失败或 Wait 返回后, 该 ctx 一定被取消.
	Context() context.Context

	// Go 启动一个子任务.panic 会被转换为错误参与失败传播.
	// 必须在 Wait 调用前调用(与 errgroup 一致).
	Go(fn func(context.Context) error)

	// Wait 阻塞直到所有子任务结束, 返回第一个非 nil 错误.
	// Wait 返回前保证作用域 ctx 已取消, 不会遗留 ctx 子节点.
	Wait() error
}

// New 基于 parent 派生一个可取消的 ctx, 并在新 goroutine 中执行 fn.
//
// parent 为 nil 时按 context.Background() 处理.
func New[T any](parent context.Context, fn func(context.Context) (T, error)) Future[T] {
	if parent == nil {
		parent = context.Background()
	}
	ctx, cancel := context.WithCancel(parent)
	f := &future[T]{
		done:   make(chan struct{}),
		cancel: cancel,
	}

	go func() {
		defer cancel()      // 最后: 无论如何释放 ctx 资源
		defer close(f.done) // 次之: 对外宣告完成(happens-before 边界)
		defer func() {
			if r := recover(); r != nil {
				f.err = fmt.Errorf("async: panic: %v\n%s", r, debug.Stack())
			}
		}()

		f.val, f.err = fn(ctx)
	}()

	return f
}

// NewScope 创建一个 Scope.parent 为 nil 时按 context.Background() 处理.
func NewScope(parent context.Context) Scope {
	if parent == nil {
		parent = context.Background()
	}
	ctx, cancel := context.WithCancel(parent)
	return &scope{
		ctx:    ctx,
		cancel: cancel,
	}
}

// All 并发执行全部任务.任一任务失败则取消其余任务并返回该错误;
// 全部成功时返回与 fns 顺序一致的结果切片.
func All[T any](ctx context.Context, fns ...func(context.Context) (T, error)) ([]T, error) {
	sc := NewScope(ctx)
	out := make([]T, len(fns))

	for i, fn := range fns {
		i, fn := i, fn // Go < 1.22 兼容; 每个 i 唯一, out[i] 无竞争
		sc.Go(func(c context.Context) error {
			v, err := fn(c)
			if err != nil {
				return err
			}
			out[i] = v
			return nil
		})
	}

	if err := sc.Wait(); err != nil {
		return nil, err
	}
	return out, nil
}

// Race 并发执行全部任务, 返回第一个完成者的结果(无论成败),
// 并通过取消通知其余任务尽快退出.
//
// 结果 channel 容量等于任务数, 失败/落败任务的发送永不阻塞;
// defer cancel 保证其余任务收到退出信号.
func Race[T any](ctx context.Context, fns ...func(context.Context) (T, error)) (T, error) {
	var zero T
	if len(fns) == 0 {
		return zero, nil
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	type result struct {
		v   T
		err error
	}
	ch := make(chan result, len(fns)) // 容量 = 任务数, loser 发送不阻塞

	for _, fn := range fns {
		fn := fn
		go func() {
			defer func() {
				if r := recover(); r != nil {
					var zv T
					ch <- result{zv, fmt.Errorf("async: race panic: %v", r)}
				}
			}()
			v, err := fn(ctx)
			ch <- result{v, err}
		}()
	}

	select {
	case r := <-ch:
		return r.v, r.err
	case <-ctx.Done():
		return zero, ctx.Err()
	}
}

// future 是 Future[T] 的唯一实现.
type future[T any] struct {
	done   chan struct{}      // 关闭即"完成"; 完成语义的唯一载体
	val    T                  // 仅当 done 已关闭后可读
	err    error              // 仅当 done 已关闭后可读
	cancel context.CancelFunc // 释放 ctx 子节点; 幂等
}

// Get 实现 Future.Get.
func (f *future[T]) Get() (T, error) {
	<-f.done
	return f.val, f.err
}

// GetContext 实现 Future.GetContext.
func (f *future[T]) GetContext(ctx context.Context) (T, error) {
	select {
	case <-f.done:
		return f.val, f.err
	case <-ctx.Done():
		var zero T
		return zero, ctx.Err()
	}
}

// TryGet 实现 Future.TryGet.
func (f *future[T]) TryGet() (val T, err error, ok bool) {
	select {
	case <-f.done:
		return f.val, f.err, true
	default:
		return val, err, false
	}
}

// Done 实现 Future.Done.
func (f *future[T]) Done() <-chan struct{} {
	return f.done
}

// Cancel 实现 Future.Cancel.
func (f *future[T]) Cancel() {
	f.cancel()
}

// scope 是 Scope 的唯一实现.
//
// 并发约定:
//   - wg 追踪所有子任务; Wait 返回后不会再启动新任务(调用方契约).
//   - mu 保护 err 与"首次失败"判定; cancel 只由首个失败者触发(幂等亦可).
type scope struct {
	ctx    context.Context    // 派生自 parent; 失败即取消
	cancel context.CancelFunc // 释放 ctx 子节点; Wait 前保证调用
	wg     sync.WaitGroup     // 子任务计数
	mu     sync.Mutex         // 保护 err 与首次失败判定
	err    error              // 第一个非 nil 错误; 无"状态"含义, 纯结果
}

// Context 实现 Scope.Context.
func (s *scope) Context() context.Context {
	return s.ctx
}

// Go 实现 Scope.Go.
func (s *scope) Go(fn func(context.Context) error) {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		defer func() {
			if r := recover(); r != nil {
				s.fail(fmt.Errorf("async: scope panic: %v\n%s", r, debug.Stack()))
			}
		}()
		if err := fn(s.ctx); err != nil {
			s.fail(err)
		}
	}()
}

// Wait 实现 Scope.Wait.
func (s *scope) Wait() error {
	s.wg.Wait()
	s.cancel() // 兜底: 即便全部成功也释放 ctx, 避免 ctx 树累积

	s.mu.Lock()
	defer s.mu.Unlock()
	return s.err
}

// fail 记录首个错误并触发取消; 后续错误忽略, 避免语义漂移.
func (s *scope) fail(err error) {
	s.mu.Lock()
	first := s.err == nil
	if first {
		s.err = err
	}
	s.mu.Unlock()

	if first {
		s.cancel()
	}
}
