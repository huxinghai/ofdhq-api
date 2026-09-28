package variable

import (
	"log"
	"ofdhq-api/app/global/my_errors"
	"ofdhq-api/app/utils/snow_flake/snowflake_interf"
	"ofdhq-api/app/utils/yml_config/ymlconfig_interf"
	"os"
	"path/filepath"
	"time"

	"github.com/casbin/casbin/v2"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// 开发者自行封装的全局变量，请做好并发安全检查与确认

var (
	BasePath                string                      // 定义项目的根目录
	EventDestroyPrefix      = "Destroy_"                //  程序退出时需要销毁的事件前缀
	ConfigKeyPrefix         = "Config_"                 //  配置文件键值缓存时，键的前缀
	DateFormatDay           = "2006-01-02"              //  设置全局日期时间格式
	DateFormat              = "2006-01-02 15:04:05"     //  设置全局日期时间格式
	DateFormatMS            = "2006-01-02 15:04:05.999" //  设置全局日期时间格式
	ShanghaiLoc             *time.Location
	TOTPIssuer              = "EZQ"
	AdminBindContextKeyName = "ADMIN_USER_ID"

	TopicCategoryOfficial       = "官方"
	TopicCategoryOfficialNotice = "平台公告"

	// 全局日志指针
	ZapLog *zap.Logger
	// 全局配置文件
	ConfigYml       ymlconfig_interf.YmlConfigInterf // 全局配置文件指针
	ConfigGormv2Yml ymlconfig_interf.YmlConfigInterf // 全局配置文件指针

	//gorm 数据库客户端，如果您操作数据库使用的是gorm，请取消以下注释，在 bootstrap>init 文件，进行初始化即可使用
	GormDbMysql      *gorm.DB // 全局gorm的客户端连接
	GormDbSqlserver  *gorm.DB // 全局gorm的客户端连接
	GormDbPostgreSql *gorm.DB // 全局gorm的客户端连接

	//雪花算法全局变量
	SnowFlake snowflake_interf.InterfaceSnowFlake

	//websocket
	WebsocketHub              interface{}
	WebsocketHandshakeSuccess = `{"code":200,"msg":"ws连接成功","data":""}`
	WebsocketServerPingMsg    = "Server->Ping->Client"

	//casbin 全局操作指针
	Enforcer *casbin.SyncedEnforcer

	//  用户自行定义其他全局变量 ↓
	DefaultAvatar = "https://ofdhq.oss-cn-shenzhen.aliyuncs.com/assets/2025_12/unnamed.jpg"

	UserRobotDeleteLockKey = "user_robot:delete"
)

func init() {
	// 1.初始化程序根目录
	if curPath, err := os.Getwd(); err == nil {
		// 路径进行处理，兼容单元测试程序启动时的奇怪路径：
		// go test 会把工作目录切到被测包目录（如 test/、app/utils/douyin/），
		// 统一向上回溯到包含 config/config.yml 的目录作为项目根
		BasePath = locateProjectRoot(curPath)
	} else {
		log.Fatal(my_errors.ErrorsBasePath)
	}
	var err error
	ShanghaiLoc, err = time.LoadLocation("Asia/Shanghai")
	if err != nil {
		log.Fatal("Error loading location:", err)
		return
	}
}

// locateProjectRoot 从 start 向上逐级查找包含 config/config.yml 的目录作为项目根，
// 找不到时原样返回 start（保持旧行为，由后续配置初始化暴露错误）
func locateProjectRoot(start string) string {
	dir := start
	for i := 0; i < 10; i++ {
		if _, err := os.Stat(filepath.Join(dir, "config", "config.yml")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return start
}

func NowTimeSH() time.Time {
	return time.Now().In(ShanghaiLoc)
}
