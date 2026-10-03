import { beforeEach, expect, test, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import axios from 'axios'
import { useHistoryStore } from './history'
import { useUserStore } from '../user'

beforeEach(() => {
  vi.restoreAllMocks()
  setActivePinia(createPinia())
})

test('远程删除使用历史记录ID，本地列表按新闻ID移除', async () => {
  useUserStore().$patch({ isLogin: true, token: 'test-token' })
  const store = useHistoryStore()
  store.history = [{ id: 11, historyId: 3 }, { id: 12, historyId: 4 }]
  const remove = vi.spyOn(axios, 'delete').mockResolvedValue({ data: { code: 200 } })
  expect(await store.removeHistoryApi(11)).toEqual({ success: true })
  expect(remove.mock.calls[0][0]).toMatch(/\/api\/history\/delete\/3$/)
  expect(store.history).toEqual([{ id: 12, historyId: 4 }])
})

test('缺少历史记录ID时不发送错误的删除请求', async () => {
  useUserStore().$patch({ isLogin: true, token: 'test-token' })
  const store = useHistoryStore()
  store.history = [{ id: 11 }]
  const remove = vi.spyOn(axios, 'delete')
  expect((await store.removeHistoryApi(11)).success).toBe(false)
  expect(remove).not.toHaveBeenCalled()
})
