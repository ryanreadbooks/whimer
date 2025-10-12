package msg

import "net/http"

func (h *Handler) GetTotalUnreadCount() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO 获取p2p未读数

		// TODO 获取系统消息未读数
	}
}
