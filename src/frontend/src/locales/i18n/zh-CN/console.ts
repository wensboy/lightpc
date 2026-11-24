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
  },
  empty: {
    title: '暂无内容',
    description: '此视图预留给控制台面板. 内容随对应功能一起落地.',
  },
}
