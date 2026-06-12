// Skill 有效性测试用例 —— 20 个核心违规代码样本 + 期望关键词断言
// 用法：eval.go 把每个用例发给 Claude API，检查回复是否命中 min_match 个关键词
package main

// TestCase 一条测试用例
type TestCase struct {
	ID            string   `json:"id"`
	Category      string   `json:"category"` // iron-law | common-lib
	Rule          string   `json:"rule"`
	Description   string   `json:"description"`
	ViolationCode string   `json:"violation_code"`
	Keywords      []string `json:"keywords"`  // AI 回复中至少要命中 min_match 个
	MinMatch      int      `json:"min_match"` // 最少命中数
}

// testCases 20 个用例，按 category 分两组
var testCases = []TestCase{
	// ============ 铁律 12 条 ============
	{
		ID:       "iron-01-hardcoded-secret",
		Category: "iron-law",
		Rule:     "铁律 1 · 禁止硬编码密钥",
		ViolationCode: `package auth

const APIKey = "sk-proj-abc123defghijklmnop"
const DBPassword = "prod-password-xyz"

func connect() {
    _ = APIKey
    _ = DBPassword
}`,
		Keywords: []string{"硬编码", "密钥", "env", "config", "泄露", "git", "铁律 1"},
		MinMatch: 3,
	},
	{
		ID:       "iron-02-bare-errors-new",
		Category: "iron-law",
		Rule:     "铁律 2 · 业务错误必须走 xerror/errno",
		ViolationCode: `func GetOrder(id int64) (*Order, error) {
    o, err := repo.Find(id)
    if err != nil {
        return nil, errors.New("order not found")
    }
    return o, nil
}`,
		Keywords: []string{"xerror", "errno", "errors.New", "错误码", "铁律 2"},
		MinMatch: 2,
	},
	{
		ID:       "iron-03-ignore-err",
		Category: "iron-law",
		Rule:     "铁律 3 · error 必检",
		ViolationCode: `func LoadConfig() *Config {
    data, _ := os.ReadFile("config.yaml")
    var c Config
    _ = json.Unmarshal(data, &c)
    return &c
}`,
		Keywords: []string{"丢弃", "忽略", "必检", "error", "铁律 3"},
		MinMatch: 2,
	},
	{
		ID:       "iron-04-business-panic",
		Category: "iron-law",
		Rule:     "铁律 4 · 业务代码禁用 panic",
		ViolationCode: `func (s *OrderService) Process(req *Req) {
    if req == nil {
        panic("req is nil")
    }
    if req.Amount < 0 {
        panic("invalid amount")
    }
}`,
		Keywords: []string{"panic", "业务", "init", "禁用", "铁律 4", "进程"},
		MinMatch: 2,
	},
	{
		ID:       "iron-05-bare-goroutine",
		Category: "iron-law",
		Rule:     "铁律 5 · goroutine 必须有退出机制",
		ViolationCode: `func StartWorker() {
    go func() {
        for {
            doWork()
            time.Sleep(time.Second)
        }
    }()
}`,
		Keywords: []string{"context", "errgroup", "退出", "泄漏", "goroutine", "铁律 5"},
		MinMatch: 2,
	},
	{
		ID:       "iron-06-float-money",
		Category: "iron-law",
		Rule:     "铁律 6 · 金额禁用 float64",
		ViolationCode: `type Order struct {
    ID     int64
    Amount float64
    Fee    float64
}

func (o *Order) Total() float64 {
    return o.Amount + o.Fee
}`,
		Keywords: []string{"decimal", "float", "精度", "金额", "铁律 6"},
		MinMatch: 2,
	},
	{
		ID:       "iron-07-local-time",
		Category: "iron-law",
		Rule:     "铁律 7 · 时间存 UTC",
		ViolationCode: `func SaveOrder(o *Order) error {
    o.CreatedAt = time.Now().Local()
    o.Timezone = "Asia/Shanghai"
    return db.Save(o).Error
}`,
		Keywords: []string{"UTC", "TIMESTAMPTZ", "时区", "本地时间", "铁律 7"},
		MinMatch: 2,
	},
	{
		ID:       "iron-08-fmt-println-log",
		Category: "iron-law",
		Rule:     "铁律 8 · 只用 slog 结构化日志",
		ViolationCode: `func HandlePayment(orderID string, amount float64) {
    fmt.Println("received payment", orderID, amount)
    log.Printf("processed: order=%s", orderID)
}`,
		Keywords: []string{"slog", "fmt.Println", "结构化", "trace_id", "铁律 8"},
		MinMatch: 2,
	},
	{
		ID:       "iron-09-select-star",
		Category: "iron-law",
		Rule:     "铁律 9 · SQL 禁止 SELECT *",
		ViolationCode: `func ListOrders(userID int64) ([]*Order, error) {
    var orders []*Order
    err := db.Raw("SELECT * FROM orders WHERE user_id = ?", userID).Scan(&orders).Error
    return orders, err
}`,
		Keywords: []string{"SELECT *", "显式", "字段", "铁律 9"},
		MinMatch: 2,
	},
	{
		ID:       "iron-10-db-foreign-key",
		Category: "iron-law",
		Rule:     "铁律 10 · 禁止数据库外键",
		ViolationCode: `CREATE TABLE trade.orders (
    id           BIGINT PRIMARY KEY,
    user_id      BIGINT NOT NULL,
    product_id   BIGINT NOT NULL,
    FOREIGN KEY (user_id) REFERENCES auth.users(id),
    FOREIGN KEY (product_id) REFERENCES catalog.products(id)
);`,
		Keywords: []string{"外键", "FOREIGN KEY", "应用层", "分库", "铁律 10"},
		MinMatch: 2,
	},
	{
		ID:       "iron-11-commented-dead-code",
		Category: "iron-law",
		Rule:     "铁律 11 · 禁止提交注释代码",
		ViolationCode: `func CancelOrder(id int64) error {
    order := findOrder(id)
    // old logic removed 2024-01
    // if order.Status == StatusShipped {
    //     return ErrCannotCancel
    // }
    // TODO: 2024-02 再看看
    order.Status = StatusCanceled
    return save(order)
}`,
		Keywords: []string{"注释", "Git", "死代码", "删除", "铁律 11"},
		MinMatch: 2,
	},
	{
		ID:       "iron-12-sensitive-in-log",
		Category: "iron-law",
		Rule:     "铁律 12 · 敏感数据禁入日志",
		ViolationCode: `func Login(ctx context.Context, phone, password, idCard string) error {
    slog.InfoContext(ctx, "user login",
        slog.String("phone", phone),
        slog.String("password", password),
        slog.String("id_card", idCard),
    )
    return nil
}`,
		Keywords: []string{"敏感", "PII", "密码", "身份证", "脱敏", "铁律 12"},
		MinMatch: 3,
	},

	// ============ common-lib 替换表 8 条 ============
	{
		ID:       "cl-01-raw-sarama",
		Category: "common-lib",
		Rule:     "裸 sarama → mask-go-common-lib/mq",
		ViolationCode: `import "github.com/IBM/sarama"

func NewOrderProducer(brokers []string) sarama.SyncProducer {
    config := sarama.NewConfig()
    config.Producer.Return.Successes = true
    p, _ := sarama.NewSyncProducer(brokers, config)
    return p
}`,
		Keywords: []string{"mask-go-common-lib", "mq", "sarama", "naming", "TopicInProject"},
		MinMatch: 2,
	},
	{
		ID:       "cl-02-raw-go-redis",
		Category: "common-lib",
		Rule:     "裸 go-redis → redisx",
		ViolationCode: `import "github.com/redis/go-redis/v9"

func GetCache(id string) (string, error) {
    client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
    key := fmt.Sprintf("order:detail:%s", id)
    return client.Get(ctx, key).Result()
}`,
		Keywords: []string{"redisx", "go-redis", "命名规范", "common-lib"},
		MinMatch: 2,
	},
	{
		ID:       "cl-03-raw-otel",
		Category: "common-lib",
		Rule:     "手工 otel.Tracer → tracing.Init",
		ViolationCode: `import "go.opentelemetry.io/otel"

func init() {
    tracer := otel.Tracer("my-service")
    _ = tracer
}`,
		Keywords: []string{"tracing.Init", "common-lib", "otel", "采样"},
		MinMatch: 2,
	},
	{
		ID:       "cl-04-hardcoded-redis-key",
		Category: "common-lib",
		Rule:     "手拼 Redis key → redisx 命名",
		ViolationCode: `func CacheOrder(id string, data []byte) error {
    key := "order_cache_" + id
    return rdb.Set(ctx, key, data, time.Hour).Err()
}`,
		Keywords: []string{"redisx", "命名", "冒号", "规范", "业务名:模块:"},
		MinMatch: 2,
	},
	{
		ID:       "cl-05-hardcoded-topic",
		Category: "common-lib",
		Rule:     "硬编码 Kafka topic → naming.TopicInProject",
		ViolationCode: `func PublishOrderPaid(evt Event) error {
    const topic = "order_paid_event"
    return producer.Send(topic, evt)
}`,
		Keywords: []string{"naming.TopicInProject", "topic", "env", "规范"},
		MinMatch: 2,
	},
	{
		ID:       "cl-06-raw-http-client",
		Category: "common-lib",
		Rule:     "裸 net/http.Client → httpclient.New",
		ViolationCode: `func CallThirdParty() (*Response, error) {
    client := &http.Client{}
    resp, err := client.Get("https://api.example.com/data")
    return parse(resp), err
}`,
		Keywords: []string{"httpclient.New", "common-lib", "超时", "span", "追踪"},
		MinMatch: 2,
	},
	{
		ID:       "cl-07-bad-package-name",
		Category: "common-lib",
		Rule:     "包名禁用 util/common/helper",
		ViolationCode: `// file: internal/util/helper.go
package util

func FormatMoney(x decimal.Decimal) string {
    return x.StringFixed(2)
}

func ValidateEmail(s string) bool {
    return strings.Contains(s, "@")
}`,
		Keywords: []string{"util", "common", "helper", "命名", "具体"},
		MinMatch: 2,
	},
	{
		ID:       "cl-08-wrong-acronym-case",
		Category: "common-lib",
		Rule:     "缩写必须全大/全小（JSONXxx 非 JsonXxx）",
		ViolationCode: `type JsonConfig struct {
    HttpUrl    string
    ApiKey     string
    JwtSecret  string
}

func NewJsonProcessor() *JsonConfig {
    return &JsonConfig{}
}`,
		Keywords: []string{"JSON", "HTTP", "API", "JWT", "缩写", "大写", "JsonConfig"},
		MinMatch: 3,
	},

	// ============ 命名规范 7 条（v1.7.37 新增）============
	{
		ID:       "naming-uid",
		Category: "naming",
		Rule:     "命名 §1.2 · 用户字段统一 uid（禁 user_id）",
		ViolationCode: `CREATE TABLE orders (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    amount NUMERIC(28,8),
    gmt_create TIMESTAMP
);`,
		Keywords: []string{"uid", "user_id", "VARCHAR(64)", "命名", "§1.2"},
		MinMatch: 2,
	},
	{
		ID:       "naming-created-at",
		Category: "naming",
		Rule:     "命名 §2.2 · 时间字段统一 _at 后缀（禁 gmt_create / create_time）",
		ViolationCode: `CREATE TABLE products (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255),
    gmt_create TIMESTAMP NOT NULL,
    gmt_modified TIMESTAMP NOT NULL
);`,
		Keywords: []string{"created_at", "updated_at", "_at", "TIMESTAMPTZ", "命名", "§2.2"},
		MinMatch: 3,
	},
	{
		ID:       "naming-soft-delete",
		Category: "naming",
		Rule:     "命名 §5.2 · 软删除统一 deleted_at（禁 is_deleted BOOLEAN）",
		ViolationCode: `CREATE TABLE accounts (
    id BIGSERIAL PRIMARY KEY,
    uid VARCHAR(64) NOT NULL,
    is_deleted BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ
);`,
		Keywords: []string{"deleted_at", "is_deleted", "TIMESTAMPTZ", "软删除", "§5.2"},
		MinMatch: 3,
	},
	{
		ID:       "naming-amount-prefix",
		Category: "naming",
		Rule:     "命名 §4.2 · 金额必带业务前缀（禁裸 amount）",
		ViolationCode: `CREATE TABLE transactions (
    id BIGSERIAL PRIMARY KEY,
    uid VARCHAR(64) NOT NULL,
    amount NUMERIC(28,8) NOT NULL,
    fee NUMERIC(28,8)
);`,
		Keywords: []string{"业务前缀", "order_amount", "fee_amount", "amount", "§4.2"},
		MinMatch: 2,
	},
	{
		ID:       "naming-bool-prefix",
		Category: "naming",
		Rule:     "命名 §5.1 · 布尔字段需 is_ / has_ / can_ 前缀",
		ViolationCode: `CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    uid VARCHAR(64) NOT NULL,
    enabled BOOLEAN,
    verified BOOLEAN,
    auto_renewal BOOLEAN
);`,
		Keywords: []string{"is_enabled", "is_verified", "is_", "前缀", "§5.1"},
		MinMatch: 3,
	},
	{
		ID:       "naming-bare-version",
		Category: "naming",
		Rule:     "命名 §6.3 · 业务版本字段必须前缀（裸 version 只能是乐观锁）",
		ViolationCode: `CREATE TABLE risk_rules (
    id BIGSERIAL PRIMARY KEY,
    rule_name VARCHAR(64),
    version INT NOT NULL,  -- 实际是风控规则版本
    content JSONB
);`,
		Keywords: []string{"rule_version", "业务版本", "乐观锁", "前缀", "§6.3"},
		MinMatch: 2,
	},
	{
		ID:       "naming-client-ip",
		Category: "naming",
		Rule:     "命名 §8 · 客户端 IP 字段统一 client_ip（禁 ip_addr / login_ip）",
		ViolationCode: `CREATE TABLE login_logs (
    id BIGSERIAL PRIMARY KEY,
    uid VARCHAR(64) NOT NULL,
    login_ip VARCHAR(45),
    ip_addr INET,
    created_at TIMESTAMPTZ
);`,
		Keywords: []string{"client_ip", "INET", "命名", "§8"},
		MinMatch: 2,
	},

	// ============ 特性开关（v1.7.37 新增）============
	{
		ID:       "ff-default-on",
		Category: "feature-flags",
		Rule:     "特性开关 §2.3 · 新 FF 默认值必须 OFF",
		ViolationCode: `// 新增上线开关
const FFEnableNewPayment = true  // 默认开

func ProcessPayment(ctx context.Context, userID string) error {
    if !FFEnableNewPayment {
        return oldPaymentService(ctx)
    }
    return newPaymentService(ctx)
}`,
		Keywords: []string{"OFF", "默认", "false", "灰度", "ff.IsEnabled", "特性开关"},
		MinMatch: 3,
	},
	{
		ID:       "ff-no-ctx",
		Category: "feature-flags",
		Rule:     "特性开关 §2.2 · 走 go-common FF 客户端（禁自实现）",
		ViolationCode: `// 自己读环境变量做开关
func IsFeatureEnabled() bool {
    return os.Getenv("ENABLE_NEW_FEATURE") == "true"
}

func handler(ctx context.Context) {
    if IsFeatureEnabled() {
        newLogic(ctx)
    } else {
        oldLogic(ctx)
    }
}`,
		Keywords: []string{"go-common", "ff.IsEnabled", "FF 客户端", "禁自实现", "IDP"},
		MinMatch: 2,
	},
	{
		ID:       "ff-commit-format",
		Category: "feature-flags",
		Rule:     "特性开关 §5 · Commit 必须带 [reqID] (@user)",
		ViolationCode: `# 提交记录
git log --oneline -5
abc123 feat: 增加钱包提现限额
def456 fix: 修复支付回调超时
ghi789 update: 优化订单列表查询`,
		Keywords: []string{"[reqID]", "@user", "Conventional Commits", "feat:", "[req"},
		MinMatch: 2,
	},
}

// modelChoices 可选模型（UI 下拉）
// 带完整日期的版本最稳；纯 alias 在某些 API 版本下会 404 model_not_found
var modelChoices = []string{
	"claude-sonnet-4-5-20250929",    // Sonnet 4.5（推荐，默认）
	"claude-sonnet-4-5",              // alias，新版本才支持
	"claude-opus-4-1-20250805",       // Opus 4.1（更强，更慢更贵）
	"claude-opus-4-1",                // alias
	"claude-3-5-sonnet-20241022",     // Sonnet 3.5（兼容性最强）
	"claude-3-5-sonnet-latest",       // alias
}
