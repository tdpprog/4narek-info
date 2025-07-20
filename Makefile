.PHONY: install
install:
	go build -o simple-backend
	useradd -r -s /bin/false simple-user || true
	mkdir -p /opt/simple-backend
	cp simple-backend /opt/simple-backend/
	cp deploy/simple-backend.service /etc/systemd/system/
	systemctl daemon-reload
	systemctl enable simple-backend
	systemctl start simple-backend

.PHONY: uninstall
uninstall:
	systemctl stop simple-backend
	systemctl disable simple-backend
	rm -rf /opt/simple-backend
	rm -f /etc/systemd/system/simple-backend.service
	systemctl daemon-reload
	userdel simple-user || true