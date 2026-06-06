import request from '@/utils/request'

export function createGame(data) {
  return request({
    url: '/games',
    method: 'post',
    data
  })
}

export function listGames(params) {
  return request({
    url: '/games',
    method: 'get',
    params
  })
}

export function getGame(id) {
  return request({
    url: `/games/${id}`,
    method: 'get'
  })
}

export function updateGame(id, data) {
  return request({
    url: `/games/${id}`,
    method: 'put',
    data
  })
}

export function deleteGame(id) {
  return request({
    url: `/games/${id}`,
    method: 'delete'
  })
}

export function parseSGF(sgf) {
  return request({
    url: '/games/parse',
    method: 'post',
    data: { sgf }
  })
}
