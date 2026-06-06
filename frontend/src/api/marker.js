import request from '@/utils/request'

export function createMarker(data) {
  return request({
    url: '/markers',
    method: 'post',
    data
  })
}

export function listMarkersByGame(gameId, params) {
  return request({
    url: `/markers/game/${gameId}`,
    method: 'get',
    params
  })
}

export function deleteMarker(id) {
  return request({
    url: `/markers/${id}`,
    method: 'delete'
  })
}
