package bee

import "time"

// Event events table struct
type Event struct {
	ID          uint64    `gorm:"column:id" db:"id" json:"id"`                               //数据表主键
	Topic       string    `gorm:"column:topic" db:"topic" json:"topic" `                     //topic 主题
	Payload     string    `gorm:"column:payload" db:"payload" json:"payload"`                //消息
	Publisher   string    `gorm:"column:publisher" db:"publisher" json:"publisher"`          //发布者
	PublishedAt time.Time `gorm:"column:published_at" db:"published_at" json:"published_at"` //发布时间 发布时间可能早于创建时间
	CreatedAt   time.Time `gorm:"column:created_at" db:"created_at" json:"created_at"`       //创建时间
	UpdatedAt   time.Time `gorm:"column:updated_at" db:"updated_at" json:"updated_at"`       //更新时间
}

// TableName return data table name
func (e *Event) TableName() string {
	return "events"
}

// Task task table model
type Task struct {
	ID         uint64    `gorm:"column:id" db:"id" json:"id" form:"id"`                                     //数据表主键
	EventID    uint64    `gorm:"column:event_id" db:"event_id" json:"event_id" form:"event_id"`             //事件id
	Topic      string    `gorm:"column:topic" db:"topic" json:"topic" form:"topic"`                         //事件主题
	Payload    string    `gorm:"column:payload" db:"payload" json:"payload" form:"payload"`                 //事件消息
	Status     uint8     `gorm:"column:status" db:"status" json:"status" form:"status"`                     //任务状态 1:就绪 2:处理中 3:结束 4:中断 5:取消
	Result     string    `gorm:"column:result" db:"result" json:"result" form:"result"`                     //任务执行结果
	StartTime  time.Time `gorm:"column:start_time" db:"start_time" json:"start_time" form:"start_time"`     //开始执行时间
	FinishTime time.Time `gorm:"column:finish_time" db:"finish_time" json:"finish_time" form:"finish_time"` //结束时间
	CreatedAt  time.Time `gorm:"column:created_at" db:"created_at" json:"created_at" form:"created_at"`     //任务创建时间
	UpdatedAt  time.Time `gorm:"column:updated_at" db:"updated_at" json:"updated_at" form:"updated_at"`     //任务更新时间
}

//TableName return data table name
func (t Task) TableName() string {
	return "tasks"
}
