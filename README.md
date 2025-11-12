# Prime API

Kullanıcı kimlik doğrulaması, oturum yönetimi, avatar yükleme ve gelecekteki içerik modülleri için hazırlanmış, Gin tabanlı bir backend iskeleti.

## Özellik Özeti
- JWT tabanlı erişim + döndürülebilir refresh token oturumları
- PostgreSQL (GORM) + Redis + MinIO/S3 ya da uyumlu obje depolama
- Zap ile yapılandırılabilir loglama, merkezi hata zarfı (Envelope) yanıtları
- Goose migrasyonları ve Swagger 2.0 (sweggo) dokümantasyonu

## Gereksinimler
- Go 1.21+ (go.mod 1.24.3 ile uyumlu)
- PostgreSQL, Redis ve bir S3/MinIO uç noktası
- `swag` ve `goose` CLI (opsiyonel ama önerilir)

## Hızlı Başlangıç
1. Depoyu klonla ve `.env` dosyasını kendi ortamına göre güncelle (`DATABASE_URL`, Redis, JWT ve S3 ayarları).
2. Migrasyonları çalıştır:
   ```bash
   make migrate-up
   ```
3. API’yi ayağa kaldır:
   ```bash
   make run
   # veya geliştirme akışı için
   make dev   # migrate-up + run
   ```
4. Varsayılan adres: `http://localhost:8080/api/v1`

## Swagger Dokümantasyonu
- Çalışırken `/docs/index.html` üzerinden UI’ye ulaşabilirsin (production’da kapalı).
- Şemayı güncellemek için:
  ```bash
  swag init -g cmd/api/main.go \
    -d cmd/api,internal \
    -o internal/apidocs \
    --parseInternal
  ```
  > Not: Swagger, Go generics içeren `common.Envelope[T]` tipini tanımıyor. Handler yorumlarında `internal/apidoc` paketindeki somut envelope tiplerini kullanmaya devam et.

## Faydalı Make Hedefleri
- `make run` – API’yi çalıştırır.
- `make build` – `cmd/api` için production derlemesi.
- `make migrate-up|down|status` – Goose migrasyon komutları.

## Proje Yapısı
- `cmd/api` – Ana HTTP uygulaması, Gin & middleware kurulumu.
- `cmd/migration` – Goose ile DB migrasyon komutları.
- `internal/app` – config, logger, DB, Redis, MinIO adapterları.
- `internal/modules` – auth ve user domain katmanları (handler, service, repo).
- `internal/apidoc[s]` – Swagger için veri zarfları ve üretilmiş dosyalar.

Katkı veya soru için issue açman yeterli; minimal ama anlaşılır tutmaya çalışıyoruz. 🚀
