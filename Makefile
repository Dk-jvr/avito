#Сброка и запуск контейнеров приложения если изменен код проекта
service-build-debug:
	sudo docker-compose build
	sudo docker-compose up

#Повторный запуск приложения если не изменен код проекта
start-service-debug:
	sudo docker-compose up
