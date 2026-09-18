# SERVER_ADDRESS=localhost \
# AUTH_SERVER_PORT=8181 \
# CRYPTO_SERVER_PORT=8080 \

# DB_USER=root \
# DB_PASSWORD=L30@sql \
# DB_HOST=localhost \
# DB_PORT=3306 \
# DB_NAME=crypto \
# go run main.go


# export PORT=8080
# export DB_HOST=localhost
# go run main.go
clear;
SERVER_ADDRESS=localhost SERVER_PORT=8181 DB_USER=root DB_PASSWORD=secure_password DB_HOST=localhost DB_PORT=3306 DB_NAME=auth go run auth/main.go