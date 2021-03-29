#!/bin/bash
sleep 20
./runBATMAN.sh 
sleep 15
cd /home/pi/apl
go run JU_led_mesh.go -rasp-id=64 --web-addr :8080 -no-duino -log-level debug
