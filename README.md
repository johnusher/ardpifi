# ardpifi

# johnusher/ardpifi Summary

Mesh network with Raspis to control Arduino controlled LED strips.

LED code can be updated, compiled and flashed locally via USB from Raspi to Arduino.

Keyboard inputs on each Raspi will send ED pattern info and sync across all Raspis in the network.

Mesh code from: https://github.com/siggy/ledmesh

## Hardware set-up
Connect Arduino Uno with Raspi 3 via USB
Run the go script, and keyboard numbers will control LED sequence on the NeoPixel strip.
LED sequence can be programmed on the Raspi, compiled using arduino-cli, and flashed from the Raspi.


## Upcoming attractions:
-integrate the mesh network to allow multiple Raspis to communicate and change the LED show, sync'd on all devices.
-accelerometer/ gyro control using I2C bus.


# thanks @siggy!

### First Boot

```bash
ssh pi@raspberrypi.local
# password: x

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



### Mesh network

See file main.go, from https://github.com/siggy/ledmesh

Based on:
https://www.reddit.com/r/darknetplan/comments/68s6jp/how_to_configure_batmanadv_on_the_raspberry_pi_3/

Step 1: Initial Setup of the Raspberry Pi 3s
...

Step 2: Install batctl
sudo apt install libnl-3-dev libnl-genl-3-dev

git clone https://git.open-mesh.org/batctl.git
cd batctl
sudo make install

Step 3: Activate and configure batman-adv
...

Step 4: Test the ad-hoc connection

Run this on both devices and note the "Cell" address. It should be the same on both devices.
...

Step 5: Test mesh communications

Run ifconfig and note the IPv4 and HWaddr assigned to wlan0 on each device...




## Code

Note 
All dependencies managed in `go.mod` now,
just add an import directive for any new depedency in your `*.go` files, and
`go run/build` should just handle it.

Install GO and source
```bash

```bash
wget https://dl.google.com/go/go1.15.6.linux-armv6l.tar.gz -O /tmp/go1.15.6.linux-armv6l.tar.gz
sudo tar -C /usr/local -xzf /tmp/go1.15.6.linux-armv6l.tar.gz
source ~/.bashrc

cat >> ~/.bashrc << 'EOF'
export GOPATH=$HOME/go
export PATH=/usr/local/go/bin:$PATH:$GOPATH/bin


EOF

source ~/.bashrc


mkdir -p ~/code/go/src/github.com/johnusher
git clone https://github.com/johnusher/ardpifi.git ~/code/go/src/github.com/


export GOPATH=$HOME/code/go/src



source ~/.bashrc
```





## Run

```bash
go run jumain.go
```

Press any key to print to screen (and eventually send to arduino).

To exit, press "q" to exit termbox, and then ctrl-c to exit the program.


# Arduino CLI install

folow instructions here:
https://siytek.com/arduino-cli-raspberry-pi/


arduino-cli core install arduino:avr

Note the directory for the Arudion project must have the same name as the main file ()

<del> 
## add libraries:
arduino-cli lib search Adafruit_NeoPixel
</del>

in duino_src:
compile:
arduino-cli compile --fqbn arduino:avr:uno duino_src

flash:
arduino-cli upload -p /dev/ttyACM0 --fqbn arduino:avr:uno duino_src

