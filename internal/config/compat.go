package config

import "fmt"

// --- Provider convenience functions (delegate to DefaultStore) ---

// GetProvider returns the config for a named provider, or nil.
func GetProvider(name string) *ProviderConfig {
	return DefaultStore().GetProvider(name)
}

// SetProvider creates or updates a provider and saves.
func SetProvider(name string, p *ProviderConfig) error {
	return DefaultStore().SetProvider(name, p)
}

// DeleteProviderByName removes a provider and its references from all profiles.
func DeleteProviderByName(name string) error {
	return DefaultStore().DeleteProvider(name)
}

// ProviderNames returns sorted provider names.
func ProviderNames() []string {
	return DefaultStore().ProviderNames()
}

// ExportProviderToEnv sets ANTHROPIC_* env vars for the named provider.
func ExportProviderToEnv(name string) error {
	return DefaultStore().ExportProviderToEnv(name)
}

// --- Profile convenience functions ---

// ReadProfileOrder returns the provider list for a profile.
func ReadProfileOrder(profile string) ([]string, error) {
	names := DefaultStore().GetProfileOrder(profile)
	if names == nil {
		return nil, fmt.Errorf("profile %q not found", profile)
	}
	return names, nil
}

// WriteProfileOrder sets the provider list for a profile.
func WriteProfileOrder(profile string, names []string) error {
	return DefaultStore().SetProfileOrder(profile, names)
}

// RemoveFromProfileOrder removes a provider from a profile.
func RemoveFromProfileOrder(profile, name string) error {
	return DefaultStore().RemoveFromProfile(profile, name)
}

// DeleteProfile deletes a profile. Cannot delete the default profile.
func DeleteProfile(profile string) error {
	return DefaultStore().DeleteProfile(profile)
}

// ListProfiles returns sorted profile names.
func ListProfiles() []string {
	return DefaultStore().ListProfiles()
}

// GetProfileConfig returns the full profile configuration.
func GetProfileConfig(profile string) *ProfileConfig {
	return DefaultStore().GetProfileConfig(profile)
}

// SetProfileConfig sets the full profile configuration.
func SetProfileConfig(profile string, pc *ProfileConfig) error {
	return DefaultStore().SetProfileConfig(profile, pc)
}

// --- Backward compatibility aliases for the "default" profile ---

// ReadFallbackOrder reads the default profile's provider order.
func ReadFallbackOrder() ([]string, error) {
	return ReadProfileOrder(DefaultStore().GetDefaultProfile())
}

// WriteFallbackOrder writes the default profile's provider order.
func WriteFallbackOrder(names []string) error {
	return WriteProfileOrder(DefaultStore().GetDefaultProfile(), names)
}

// RemoveFromFallbackOrder removes a provider from the default profile.
func RemoveFromFallbackOrder(name string) error {
	return RemoveFromProfileOrder(DefaultStore().GetDefaultProfile(), name)
}

// --- Global Settings convenience functions ---

// GetDefaultProfile returns the configured default profile name.
func GetDefaultProfile() string {
	return DefaultStore().GetDefaultProfile()
}

// SetDefaultProfile sets the default profile name.
func SetDefaultProfile(profile string) error {
	return DefaultStore().SetDefaultProfile(profile)
}

// GetDefaultClient returns the configured default client.
func GetDefaultClient() string {
	return DefaultStore().GetDefaultClient()
}

// SetDefaultClient sets the default client.
func SetDefaultClient(client string) error {
	return DefaultStore().SetDefaultClient(client)
}

// GetWebPort returns the configured web UI port.
func GetWebPort() int {
	return DefaultStore().GetWebPort()
}

// SetWebPort sets the web UI port.
func SetWebPort(port int) error {
	return DefaultStore().SetWebPort(port)
}

// GetProxyPort returns the configured proxy port.
func GetProxyPort() int {
	return DefaultStore().GetProxyPort()
}

// SetProxyPort sets the proxy port.
func SetProxyPort(port int) error {
	return DefaultStore().SetProxyPort(port)
}

// --- Project Bindings convenience functions ---

// BindProject binds a directory path to a profile and/or CLI.
func BindProject(path string, profile string, cli string) error {
	return DefaultStore().BindProject(path, profile, cli)
}

// UnbindProject removes the binding for a directory path.
func UnbindProject(path string) error {
	return DefaultStore().UnbindProject(path)
}

// GetProjectBinding returns the binding for a directory path.
func GetProjectBinding(path string) *ProjectBinding {
	return DefaultStore().GetProjectBinding(path)
}

// GetAllProjectBindings returns all project bindings.
func GetAllProjectBindings() map[string]*ProjectBinding {
	return DefaultStore().GetAllProjectBindings()
}

// --- Web Password convenience functions ---

// GetWebPasswordHash returns the stored bcrypt password hash.
func GetWebPasswordHash() string {
	return DefaultStore().GetWebPasswordHash()
}

// SetWebPasswordHash sets the bcrypt password hash.
func SetWebPasswordHash(hash string) error {
	return DefaultStore().SetWebPasswordHash(hash)
}

// --- Sync Config convenience functions ---

// GetSyncConfig returns the sync configuration, or nil if not configured.
func GetSyncConfig() *SyncConfig {
	return DefaultStore().GetSyncConfig()
}

// SetSyncConfig sets the sync configuration.
func SetSyncConfig(cfg *SyncConfig) error {
	return DefaultStore().SetSyncConfig(cfg)
}

// --- Pricing convenience functions ---

// GetPricing returns the model pricing map (custom overrides merged with defaults).
func GetPricing() map[string]*ModelPricing {
	return DefaultStore().GetPricing()
}

// SetPricing sets custom model pricing overrides.
func SetPricing(pricing map[string]*ModelPricing) error {
	return DefaultStore().SetPricing(pricing)
}

// --- Budget convenience functions ---

// GetBudgets returns the budget configuration.
func GetBudgets() *BudgetConfig {
	return DefaultStore().GetBudgets()
}

// SetBudgets sets the budget configuration.
func SetBudgets(budgets *BudgetConfig) error {
	return DefaultStore().SetBudgets(budgets)
}

// --- Webhook convenience functions ---

// GetWebhooks returns all webhook configurations.
func GetWebhooks() []*WebhookConfig {
	return DefaultStore().GetWebhooks()
}

// SetWebhooks sets all webhook configurations.
func SetWebhooks(webhooks []*WebhookConfig) error {
	return DefaultStore().SetWebhooks(webhooks)
}

// GetWebhook returns a webhook by name.
func GetWebhook(name string) *WebhookConfig {
	return DefaultStore().GetWebhook(name)
}

// AddWebhook adds or updates a webhook configuration.
func AddWebhook(webhook *WebhookConfig) error {
	return DefaultStore().AddWebhook(webhook)
}

// DeleteWebhook removes a webhook by name.
func DeleteWebhook(name string) error {
	return DefaultStore().DeleteWebhook(name)
}

// --- Health Check convenience functions ---

// GetHealthCheck returns the health check configuration.
func GetHealthCheck() *HealthCheckConfig {
	return DefaultStore().GetHealthCheck()
}

// SetHealthCheck sets the health check configuration.
func SetHealthCheck(hc *HealthCheckConfig) error {
	return DefaultStore().SetHealthCheck(hc)
}

// --- Compression convenience functions (BETA) ---

// GetCompression returns the compression configuration.
func GetCompression() *CompressionConfig {
	return DefaultStore().GetCompression()
}

// SetCompression sets the compression configuration.
func SetCompression(cc *CompressionConfig) error {
	return DefaultStore().SetCompression(cc)
}

// --- Middleware convenience functions (BETA) ---

// GetMiddleware returns the middleware configuration.
func GetMiddleware() *MiddlewareConfig {
	return DefaultStore().GetMiddleware()
}

// SetMiddleware sets the middleware configuration.
func SetMiddleware(mc *MiddlewareConfig) error {
	return DefaultStore().SetMiddleware(mc)
}

// --- Agent convenience functions (BETA) ---

// GetAgent returns the agent configuration.
func GetAgent() *AgentConfig {
	return DefaultStore().GetAgent()
}

// SetAgent sets the agent configuration.
func SetAgent(ac *AgentConfig) error {
	return DefaultStore().SetAgent(ac)
}

// --- Bot convenience functions ---

// GetBot returns the bot configuration.
func GetBot() *BotConfig {
	return DefaultStore().GetBot()
}

// SetBot sets the bot configuration.
func SetBot(bc *BotConfig) error {
	return DefaultStore().SetBot(bc)
}
