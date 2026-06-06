import request from '@/utils/request'

export function createProblem(data) {
  return request({
    url: '/problems',
    method: 'post',
    data
  })
}

export function listProblems(params) {
  return request({
    url: '/problems',
    method: 'get',
    params
  })
}

export function getProblem(id) {
  return request({
    url: `/problems/${id}`,
    method: 'get'
  })
}

export function updateProblem(id, data) {
  return request({
    url: `/problems/${id}`,
    method: 'put',
    data
  })
}

export function deleteProblem(id) {
  return request({
    url: `/problems/${id}`,
    method: 'delete'
  })
}

export function attemptProblem(id, data) {
  return request({
    url: `/problems/${id}/attempt`,
    method: 'post',
    data
  })
}

export function listProblemAttempts(id) {
  return request({
    url: `/problems/${id}/attempts`,
    method: 'get'
  })
}
