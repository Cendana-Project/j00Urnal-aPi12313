# 📧 SendGrid Configuration untuk Render

## Step 1: Setup SendGrid API Key

1. Login ke [SendGrid Dashboard](https://app.sendgrid.com)
2. Buka **Settings** → **API Keys**
3. Klik **"Create API Key"**
4. Pilih **"Full Access"** atau **"Restricted Access"** dengan permissions:
   - ✅ Mail Send
5. Copy API Key (misalnya: `SG.xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx`)

## Step 2: Environment Variables untuk Render

Masukkan environment variables berikut di **Render Dashboard** → **Web Service** → **Environment**:

```bash
# SMTP Configuration - SendGrid
SMTP_HOST=smtp.sendgrid.net
SMTP_PORT=587
SMTP_USERNAME=apikey
SMTP_PASSWORD=SG.xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
SMTP_FROM=no-reply@medikaone.id
```

**Catatan:**
- `SMTP_USERNAME` harus **"apikey"** (untuk SendGrid)
- `SMTP_PASSWORD` adalah **SendGrid API Key** Anda
- `SMTP_FROM` adalah email sender (harus verified di SendGrid)

## Step 3: Verify Sender Email di SendGrid

1. Buka **Settings** → **Sender Authentication**
2. Klik **"Verify a Single Sender"**
3. Isi email yang akan digunakan (misalnya: no-reply@medikaone.id)
4. Verifikasi via email yang dikirim SendGrid

## Step 4: Deploy & Test

Setelah environment variables di-set di Render, aplikasi akan auto-redeploy.

### Test Endpoints:

```bash
# Test Register (akan kirim email PIN)
curl -X POST https://medikaone-api.onrender.com/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "username": "testuser",
    "password": "Test123!@#",
    "phone": "081234567890"
  }'

# Test Forgot Password (akan kirim email PIN)
curl -X POST https://medikaone-api.onrender.com/v1/auth/forgot-password \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com"
  }'
```

## Troubleshooting

### Error: "Failed to send email"
- ✅ Pastikan SendGrid API Key valid
- ✅ Pastikan SMTP_USERNAME="apikey"
- ✅ Check SendGrid dashboard untuk email logs
- ✅ Pastikan sender email sudah verified

### Email tidak sampai
- ✅ Check spam folder
- ✅ Verifikasi sender email di SendGrid
- ✅ Check SendGrid Activity Feed untuk status delivery

## Kode sudah support SendGrid! 🎉

Kode SMTP Anda sudah generic dan siap untuk SendGrid, Mailgun, atau provider SMTP lainnya.
Perbedaan hanya di environment variables saja.

## SendGrid Free Tier:
- ✅ 100 emails/day free forever
- ✅ Detailed analytics
- ✅ Email delivery tracking
- ✅ No credit card required

Untuk kebutuhan lebih dari 100 emails/day, upgrade ke paid plan.
