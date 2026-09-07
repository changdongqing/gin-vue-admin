/**
 * @antv/g2 可选 peer 本地桩（S2 的延伸图形能力，本报表平台不使用）：
 * 仅导出空实现，避免引入约 1MB 的 g2 依赖。配合 vite alias '@antv/g2' 指向本文件。
 */
export const corelib = () => ({})

export const renders = {}

export async function renderToMountedElement() {
  /* 空实现：分析报表不使用 g2 渲染 */
}

export default { corelib, renders, renderToMountedElement }
