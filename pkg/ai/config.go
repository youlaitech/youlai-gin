package ai

// Config AI 配置
type Config struct {
	BaseURL   string `mapstructure:"baseUrl"`
	APIKey    string `mapstructure:"apiKey"`
	Model     string `mapstructure:"model"`
	TimeoutMs int    `mapstructure:"timeoutMs"`
	Provider  string `mapstructure:"provider"`
}
