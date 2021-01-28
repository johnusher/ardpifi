# ardpifi

# johnusher/ardpifi Summary

Mesh network with Raspis to control Arduino controlled LED strips using GPS and accelerometers.

A  webserver on the pi provides a bi-directional UI (currently only via ethernet SSH).

LED code can be updated, compiled and flashed locally via USB from Raspi to Arduino.

Keyboard inputs on each Raspi will send LED pattern info and sync across all Raspis in the network.

We install everything now with /bin/bootstrap (not tested!!)

Mesh code from: https://github.com/siggy/ledmesh

## Hardware set-up
Connect Arduino Uno (also Nano clone tested) to Raspi 3 via USB. Programmable NeoPixelLED stips connected to Arduino and should be powered seperately:
https://www.christians-shop.de/Nano-V3-developer-board-for-Arduino-IDE-ATMEL-ATmega328P-AVR-Microcontroller-CH340-Chip-Christians-Technikshop

NEO-6M GPS module connect with GPIO serial (uses /dev/ttyS0, you need to disable serial console output, and disable bluetooth.):
https://www.amazon.de/dp/B088LR3488/ref=pe_3044161_185740101_TE_item

sudo raspi-config, select interfacing options > serial to disable shell, ebnable hardware, reboot

LCD 1602 I2C Module | 16x2 conected via I2C: https://www.christians-shop.de/Set-LCD-1602-I2C-Module-16x2-Figures-Illumination-Blue-I2C-Module-for-Arduino

There are "no harware" options for running without LCD or GPS.

Run the go script, and keyboard numbers will send a command to the mesh network, eg controlling LED sequence on the LED strip.

LED sequences can be programmed on the Raspi, compiled using arduino-cli, and flashed from the Raspi.


## Upcoming attractions:
-send GPS on the mesh so devices can work out relative baring (using magnetometer compass).
-accelerometer/ gyro control using I2C bus.


# thanks @siggy!

### OS

1. Download Raspian Lite: https://downloads.raspberrypi.org/raspbian_lite_latest
2. Flash `20XX-XX-XX-raspbian-stretch-lite.zip` using Etcher
3. Remove/reinsert flash drive
4. Add `ssh` and `bootstrap` files:
    ```bash
    touch /Volumes/boot/ssh
    cp bin/bootstrap /Volumes/boot/
    chmod a+x /Volumes/boot/bootstrap
    diskutil umount /Volumes/boot
    ```

### First Boot

```bash
ssh pi@raspberrypi.local
# password: raspberry

# change default password
passwd



# set quiet boot
sudo sed -i '${s/$/ quiet loglevel=1/}' /boot/cmdline.txt

# install packages
sudo apt-get update
sudo apt-get install -y git tmux vim dnsmasq hostapd

# set up wifi (note leading space to avoid bash history)
sudo tee --append /etc/wpa_supplicant/wpa_supplicant.conf > /dev/null << 'EOF'
network={
    ssid="<WIFI_SSID>"
    psk="<WIFI_PASSWORD>"
}
EOF

# set static IP address
sudo tee --append /etc/dhcpcd.conf > /dev/null << 'EOF'

# set static ip

interface eth0
static ip_address=192.168.1.164/24
static routers=192.168.1.1
static domain_name_servers=192.168.1.1

interface wlan0
static ip_address=192.168.1.164/24
static routers=192.168.1.1
static domain_name_servers=192.168.1.1
EOF

# reboot to connect over wifi
sudo shutdown -r now
```

```bash
# configure git
git config --global push.default simple
git config --global core.editor "vim"
git config --global user.email "you@example.com"
git config --global user.name "Your Name"

# disable services
sudo systemctl disable hciuart
sudo systemctl disable bluetooth
sudo systemctl disable plymouth

# remove unnecessary packages
sudo apt-get -y purge libx11-6 libgtk-3-common xkb-data lxde-icon-theme raspberrypi-artwork penguinspuzzle ntp plymouth*
sudo apt-get -y autoremove

sudo raspi-config nonint do_boot_behaviour B2 0
sudo raspi-config nonint do_boot_wait 1
sudo raspi-config nonint do_serial 1
```

# now you are setup! run bootstrap to download packages

/bootstrap
```

## Build and run (in ~/code/go/src/github.com/siggy/ledmesh)

```bash
go run JU_led_mesh.go
```


### Mesh network

On raspi #1 run

./runBATMAN.sh

On raspi #2 run

./runBATMAN2.sh

NB you may need to run these commands twice!

See file main.go, from https://github.com/siggy/ledmesh

Based on:

https://www.reddit.com/r/darknetplan/comments/68s6jp/how_to_configure_batmanadv_on_the_raspberry_pi_3/

see install instructions here:

https://github.com/siggy/ledmesh/blob/master/bin/bootstrap

eg

sudo ifconfig bat0 172.27.0.1/16

and sudo ifconfig bat0 172.27.0.2/16


## Code

Note
All dependencies managed in `go.mod` now,
just add an import directive for any new depedency in your `*.go` files, and
`go run/build` should just handle it.


All code (and go) is installed via bootstrap.



# Arduino CLI install

folow instructions here:
https://siytek.com/arduino-cli-raspberry-pi/

this additional command was needed:
arduino-cli core install arduino:avr

Note the directory for the Arudio project must have the same name as the main file .ino.

Tested with Aurdion Uno and Aurdion clone: "Nano V3 | ATMEL ATmega328P AVR Microcontroller | CH340-Chip".
The Uno shows on port ttyACM0 and the clone on ttyUSB.
NB only 1 from 2 clones works for me!


## add libraries:

arduino-cli lib search Adafruit_NeoPixel


in duino_src:

compile and flash:
Uno:
arduino-cli compile --fqbn arduino:avr:uno duino_src

arduino-cli upload -p /dev/ttyACM0 --fqbn arduino:avr:uno duino_src

Clone

arduino-cli compile --fqbn arduino:avr:diecimila:cpu=atmega328 duino_src

arduino-cli upload -p /dev/ttyUSB0 --fqbn arduino:avr:diecimila:cpu=atmega328 duino_src

arduino-cli upload -p /dev/ttyACM0 --fqbn arduino:avr:diecimila:cpu=atmega328 duino_src


## Run


On raspi #1 run

```bash
./runBATMAN.sh
go run JU_led_mesh.go -rasp-id=wee --web-addr :8080 -no-lcd -log-level debug
```

On raspi #2 run

```bash
./runBATMAN2.sh
go run JU_led_mesh.go -rasp-id=poo --web-addr :8081 -no-lcd -log-level debug
```

Press any key, sent to mesh, and if it is a 0 or 1, we change led pattern

To exit, press "q" to exit termbox, and then ctrl-c to exit the program.

### Flags

```bash
$ go run JU_led_mesh.go -h
  -log-level string
    	log level, must be one of: panic, fatal, error, warn, info, debug, trace (default "info")
  -no-batman
    	run without batman network
  -no-duino
    	run without arduino
  -no-gps
    	run without gps
  -no-lcd
    	run without lcd display
  -rasp-id string
    	unique raspberry pi ID (default "raspi 1")
  -web-addr string
    	address to serve web on (default ":8080")
```

### Run without hardware

Run with hardware (serial, network) API calls mocked out:

```bash
go run JU_led_mesh.go --rasp-id "raspi 1" --web-addr :8080 --no-batman --no-duino --no-gps --no-lcd --log-level debug
```

# Set up port forwarding for web server

## First-time ssh config setup

On laptop:


```bash
PI_IP=192.168.1.164
USER=pi

cat << EOF >> ~/.ssh/config

Host pi
  HostName $PI_IP
  User $USER
  Port 22
  BatchMode yes
  ServerAliveInterval 60
  ServerAliveCountMax 5
  ForwardAgent yes
    LocalForward 127.0.0.1:8080 127.0.0.1:8080
EOF
```

## Tunnel

```bash
ssh $USER@pi
```

Open a browser window to localhost:8080
