# App Store Connect API

Then you need to generate an API Key from the App Store Connect portal (https://developer.apple.com/documentation/appstoreconnectapi/creating_api_keys_for_app_store_connect_api).

## Config file

create file `config.json`

example
```json
{
    "kid": "2X9R4HXF34",
    "iis": "57246542-96fe-1a63-e053-0824d011072a",
    "p8key": "-----BEGIN PRIVATE KEY-----\nMIGTAgEAMBMGByqGSM49AgEGCCqGSM49AwEHBHkwdwIBAQQg8/TApNI40Mz5ydZk\nDm+cem1iqGCHonQwf/NoRD4rR86gCgYIKoZIzj0DAQehRANCAAS8Vy1oF/LmqRyq\niyujDmTZ19G1B1JtCPDNPI5nlVx9kd0NwvZLfJDy9vb9nqqMTE6BZ6WLysmxW9cZ\n09N3Rcck\n-----END PRIVATE KEY-----"
}
```

| Fieild | Name | Description |
---------|------|------------
| kid | Key Identifier | Your private key ID from App Store Connect; for example 2X9R4HXF34. |
| iss | Issuer ID | Your issuer ID from the API Keys page in App Store Connect; for example, 57246542-96fe-1a63-e053-0824d011072a. |
| p8key | Private key | Download private key from App Store Connect; https://appstoreconnect.apple.com/access/integrations/api |
