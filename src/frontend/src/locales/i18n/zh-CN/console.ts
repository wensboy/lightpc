// Console holds the text of the console sheet: its navigation labels, the log panel and the empty state.
export const consoleText = {
  title: '控制台',
  description: '在一处查看应用的运行状态.',
  sections: {
    log: '日志',
    performance: '性能',
  },
  log: {
    listener: '日志监听器',
    filter: {
      level: '级别',
      pattern: '筛选（如 text, excludeText…）',
      patternLabel: '筛选日志',
    },
    levels: {
      debug: 'DEBUG',
      info: 'INFO',
      warn: 'WARN',
      error: 'ERROR',
    },

    actions: {
      clear: '清空日志',
      autoScroll: '自动滚动',
    },
    // 每个面板拥有自己的空状态: 日志说明自己的过滤器, 尚未实现的面板用占位文案. 共用一份会让除第一个面板外的
    // 文案都不成立.
    empty: {
      title: '暂无日志',
      description: '后端写入的记录会显示在这里. 如果预期有输出, 请切换来源或清除过滤器.',
    },
  },
  pending: {
    title: '暂无内容',
    description: '此面板随对应功能一起落地.',
  },
}
