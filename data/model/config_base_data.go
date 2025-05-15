package model

type BaseConfig struct {
	BaseUrl                string   `json:"base_url"`
	BasePath               string   `json:"base_path"`
	ServerPort             string   `json:"server_port"`
	PasetoAsymmetricSecret string   `json:"paseto_asymm_secret"`
	PasetoAsymmetricPublic string   `json:"paseto_asymm_public"`
	AESGCMKey              string   `json:"aes_gcm_key"`
	FileUploadExtension    []string `json:"file_upload_extension"`
	EncryptionBehavior     string   `json:"encryption_behavior"`
	MidtransMerchantID     string   `json:"midtrans_merchant_id"`
	MidtransServerKey      string   `json:"midtrans_server_key"`
	MidtransClientKey      string   `json:"midtrans_client_key"`
	CloudflareAPIKey       string   `json:"cloudflare_api_key"`
	CloudflareEmail        string   `json:"cloudflare_email"`
	CloudflareAccountID    string   `json:"cloudflare_account_id"`
	CloudflareTunnelID     string   `json:"cloudflare_tunnel_id"`
}
