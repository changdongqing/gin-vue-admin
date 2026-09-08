import service from '@/utils/request'

// @Router /post/createPost [post]
export const createPost = (data) => {
  return service({
    url: '/post/createPost',
    method: 'post',
    data
  })
}

// @Router /post/updatePost [put]
export const updatePost = (data) => {
  return service({
    url: '/post/updatePost',
    method: 'put',
    data
  })
}

// @Router /post/deletePost [delete]
export const deletePost = (params) => {
  return service({
    url: '/post/deletePost',
    method: 'delete',
    params
  })
}

// @Router /post/findPost [get]
export const findPost = (params) => {
  return service({
    url: '/post/findPost',
    method: 'get',
    params
  })
}

// @Router /post/getPostList [post]
export const getPostList = (data) => {
  return service({
    url: '/post/getPostList',
    method: 'post',
    data
  })
}

// @Router /post/getPostListAll [get]
export const getPostListAll = () => {
  return service({
    url: '/post/getPostListAll',
    method: 'get'
  })
}

// @Router /post/getPostUsers [get]
export const getPostUsers = (params) => {
  return service({
    url: '/post/getPostUsers',
    method: 'get',
    params
  })
}

// @Router /post/setPostUsers [post]
export const setPostUsers = (data) => {
  return service({
    url: '/post/setPostUsers',
    method: 'post',
    data
  })
}

// @Router /post/getUserPosts [get]
export const getUserPosts = (params) => {
  return service({
    url: '/post/getUserPosts',
    method: 'get',
    params
  })
}

// @Router /post/setUserPosts [post]
export const setUserPosts = (data) => {
  return service({
    url: '/post/setUserPosts',
    method: 'post',
    data
  })
}
