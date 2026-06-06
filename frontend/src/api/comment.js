import request from '@/utils/request'

export function createComment(data) {
  return request({
    url: '/comments',
    method: 'post',
    data
  })
}

export function listCommentsByGame(gameId, params) {
  return request({
    url: `/comments/game/${gameId}`,
    method: 'get',
    params
  })
}

export function updateComment(id, data) {
  return request({
    url: `/comments/${id}`,
    method: 'put',
    data
  })
}

export function deleteComment(id) {
  return request({
    url: `/comments/${id}`,
    method: 'delete'
  })
}
