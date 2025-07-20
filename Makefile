.PHONY: install
install:
	go build -o backend
	useradd -r -s /bin/false simple-user || true
	mkdir -p /opt/backend
	cp simple-backend /opt/backend/
	cp deploy/backend.service /etc/systemd/system/
	systemctl daemon-reload
	systemctl enable backend
	systemctl start backend

.PHONY: uninstall
uninstall:
	systemctl stop backend
	systemctl disable backend
	rm -rf /opt/backend
	rm -f /etc/systemd/system/backend.service
	systemctl daemon-reload
	userdel simple-user || true