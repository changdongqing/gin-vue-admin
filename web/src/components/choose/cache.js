// 通用选择器数据层（选用户/选部门唯一对接后端的位置）
// 模块级 reactive 缓存 + TTL + 单飞（并发调用只发一次请求），
// 范式同 gvaGrid/renderers.js 的 dictCache——缓存异步到位后，依赖它的组件/渲染器自动更新
import { reactive } from 'vue'
import { getUserSimpleList } from '@/api/user'
import { getDepartmentList } from '@/api/department'
import { getPostListAll } from '@/api/post'

const CACHE_TTL = 5 * 60 * 1000 // 5 分钟（select/单元格场景的新鲜度上限，dialog 打开会强刷）

const state = reactive({
  users: [], // UserSimple[]（后端已过滤冻结用户）
  userMap: {}, // id -> UserSimple
  departments: [], // 平铺 SysDepartment[]（含停用——树中可见不可选）
  deptMap: {}, // id -> SysDepartment
  deptTree: [], // getDepartmentList 原样树
  posts: [], // 启用岗位 SysPost[]
  loadedAt: { users: 0, departments: 0, posts: 0 },
  inflight: {} // 进行中的 Promise（单飞去重）
})

// 组件/渲染器从 reactive 缓存读取（computed 内取值，缓存到位自动更新）
export { state }

function isFresh(key) {
  return state.loadedAt[key] !== 0 && Date.now() - state.loadedAt[key] < CACHE_TTL
}

// 单飞：同 key 进行中的请求直接复用（force 与否共用一个 inflight，await 它即得新数据）
function runOnce(key, loader) {
  if (state.inflight[key]) return state.inflight[key]
  const p = Promise.resolve()
    .then(loader)
    .finally(() => {
      delete state.inflight[key]
    })
  state.inflight[key] = p
  return p
}

// 用户全量缓存（force 供弹窗打开时强刷，人员调动尽快可见）
export function ensureUsers(force = false) {
  if (!force && isFresh('users')) return Promise.resolve(state.users)
  return runOnce('users', async () => {
    const res = await getUserSimpleList()
    const list = res.data.list || []
    state.users = list
    const map = {}
    list.forEach((u) => {
      map[u.ID] = u
    })
    state.userMap = map
    state.loadedAt.users = Date.now()
    return list
  })
}

// 部门树缓存（同时维护平铺表与 id 映射，供子树过滤/名称解析）
export function ensureDepts(force = false) {
  if (!force && isFresh('departments')) return Promise.resolve(state.deptTree)
  return runOnce('departments', async () => {
    const res = await getDepartmentList()
    const tree = res.data.list || []
    state.deptTree = tree
    const flat = []
    const map = {}
    const walk = (nodes) => {
      ;(nodes || []).forEach((n) => {
        flat.push(n)
        map[n.ID] = n
        walk(n.children)
      })
    }
    walk(tree)
    state.departments = flat
    state.deptMap = map
    state.loadedAt.departments = Date.now()
    return tree
  })
}

// 启用岗位缓存（getPostListAll 仅返回启用岗位，供 dialog 筛选项与行内岗位 tag）
export function ensurePosts(force = false) {
  if (!force && isFresh('posts')) return Promise.resolve(state.posts)
  return runOnce('posts', async () => {
    const res = await getPostListAll()
    state.posts = res.data.list || []
    state.loadedAt.posts = Date.now()
    return state.posts
  })
}

// 部门后代集合（含自身）：物化路径 parentIds "0,1,3," 的前缀匹配，
// 与后端 GetDepartmentSubTreeIds 的 selfPath = parentIds + ID + "," 语义对齐
export function deptDescendantIds(departmentId) {
  const ids = new Set([Number(departmentId)])
  const dept = state.deptMap[Number(departmentId)]
  if (!dept) return ids
  const selfPath = `${dept.parentIds || ''}${dept.ID},`
  state.departments.forEach((d) => {
    if (Number(d.ID) !== Number(dept.ID) && (d.parentIds || '').startsWith(selfPath)) {
      ids.add(Number(d.ID))
    }
  })
  return ids
}

// 用户三维过滤（dialog 三维联动与 select 形态 params 预过滤共用；条件为 AND，岗位为并集语义）
export function filterUsers({ keyword, departmentId, postIds } = {}) {
  const kw = (keyword || '').trim().toLowerCase()
  const deptIds = departmentId ? deptDescendantIds(departmentId) : null
  const pidSet = postIds && postIds.length ? new Set(postIds.map(Number)) : null
  return state.users.filter((u) => {
    if (kw) {
      const nick = (u.nickName || '').toLowerCase()
      const name = (u.userName || '').toLowerCase()
      if (!nick.includes(kw) && !name.includes(kw)) return false
    }
    if (deptIds && !deptIds.has(Number(u.departmentId))) return false
    if (pidSet && !(u.postIds || []).some((id) => pidSet.has(Number(id)))) return false
    return true
  })
}

// ID → 显示名（单元格渲染、纯显示场景；未命中返回 ''——缓存未就绪由调用方回退显示原值）
export function resolveUserLabel(id) {
  const u = state.userMap[Number(id)]
  if (!u) return ''
  return u.nickName ? `${u.nickName}（${u.userName}）` : u.userName
}

export function resolveDeptLabel(id) {
  const d = state.deptMap[Number(id)]
  return d ? d.name : ''
}

// 岗位ID组 → 岗位名组（dialog 用户行内岗位 tag）
export function resolvePostLabels(postIds) {
  if (!postIds || !postIds.length) return []
  const map = {}
  state.posts.forEach((p) => {
    map[p.ID] = p.postName
  })
  return postIds.map((id) => map[Number(id)]).filter(Boolean)
}
