#!/bin/sh
set -e

DOMAIN="sopeko.com"
EMAIL="stuneak@gmail.com"
CERT_PATH="/etc/letsencrypt/live/$DOMAIN"

# Check if certificates already exist and are valid
if [ -d "$CERT_PATH" ] && [ -f "$CERT_PATH/fullchain.pem" ] && [ -f "$CERT_PATH/privkey.pem" ]; then
    echo "SSL certificates already exist for $DOMAIN"
    echo "Checking certificate validity..."

    # Check if cert is valid and not expiring within 30 days
    if openssl x509 -checkend 2592000 -noout -in "$CERT_PATH/fullchain.pem" 2>/dev/null; then
        echo "Certificate is valid and not expiring within 30 days"
        exit 0
    else
        echo "Certificate is expiring soon or invalid, will attempt renewal"
    fi
fi

echo "Obtaining SSL certificate for $DOMAIN..."

# Create webroot directory
mkdir -p /var/www/certbot

# Use standalone mode for initial certificate acquisition
# This runs certbot's own HTTP server on port 80
# Subsequent renewals will use webroot mode (handled by certbot container)
certbot certonly \
    --standalone \
    --email "$EMAIL" \
    --agree-tos \
    --no-eff-email \
    --non-interactive \
    --domains "$DOMAIN" \
    --domains "www.$DOMAIN" \
    --preferred-challenges http

# Configure renewal to use webroot mode and renew at 60 days (2 months)
# Update the renewal config
RENEWAL_CONF="/etc/letsencrypt/renewal/$DOMAIN.conf"
if [ -f "$RENEWAL_CONF" ]; then
    # Change authenticator from standalone to webroot for renewals
    sed -i 's/authenticator = standalone/authenticator = webroot/' "$RENEWAL_CONF"
    # Add webroot path if not present
    if ! grep -q "webroot_path" "$RENEWAL_CONF"; then
        echo "webroot_path = /var/www/certbot," >> "$RENEWAL_CONF"
    fi
    # Set renewal to happen at 60 days before expiry (roughly 30 days after issuance)
    # Note: Let's Encrypt certs are valid for 90 days
    if ! grep -q "renew_before_expiry" "$RENEWAL_CONF"; then
        echo "renew_before_expiry = 60 days" >> "$RENEWAL_CONF"
    fi
fi

echo "SSL certificate obtained successfully!"
echo "Certificate path: $CERT_PATH"
echo "Renewal configured for 60 days before expiry (approximately every 2 months)"
