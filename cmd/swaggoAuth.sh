cd auth
clear
swag fmt
swag init -g app/app.go --parseDependency --parseInternal
cd ..