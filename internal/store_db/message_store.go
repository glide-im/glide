package store_db

import (
	"database/sql"
	"fmt"
	"github.com/glide-im/glide/internal/config"
	"github.com/glide-im/glide/pkg/messages"
	_ "github.com/go-sql-driver/mysql"
	"time"
)

type ChatMessageStore struct {
	db *sql.DB
}

func New(conf *config.MySqlConf) (*ChatMessageStore, error) {
	db := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", conf.Username, conf.Password, conf.Host, conf.Port, conf.Db)
	open, err := sql.Open("mysql", db)
	if err != nil {
		return nil, err
	}
	m := &ChatMessageStore{
		db: open,
	}
	return m, nil
}

func (D *ChatMessageStore) StoreMessage(m *messages.ChatMessage) error {

	lg := m.From
	sm := m.To
	if lg < sm {
		lg, sm = sm, lg
	}
	sid := lg + "_" + sm

	// mysql only
	_, err := D.db.Exec(
		"INSERT INTO im_chat_message (`m_id`, `session_id`, `from`, `to`, `type`, `content`, `send_at`, `create_at`, `cli_seq`, `status`) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)ON DUPLICATE KEY UPDATE send_at=?",
		m.Mid, sid, m.From, m.To, m.Type, m.Content, m.SendAt, time.Now().Unix(), 0, 0,
		m.SendAt)
	return err
}
