# ardpifi

# johnusher/ardpifi

Connect Arduino Uno with Raspi 3 via USB.
Run the go script, and keyboard numbers will control LED sequence on the NeoPixel strip.
LED sequence can be programmed on the Raspi, compiled using arduino-cli, and flashed from the Raspi.

upcoming attractions:
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
static ip_address=192.168.1.141/24
static routers=192.168.1.1
static domain_name_servers=192.168.1.1

interface wlan0
static ip_address=192.168.1.142/24
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

## Code

Note 
All dependencies managed in `go.mod` now,
just add an import directive for any new depedency in your `*.go` files, and
`go run/build` should just handle it.

```bash



url='https://golang.org'$(curl https://golang.org/dl/ | grep armv6l | sort --version-sort | tail -1 | grep -o -E "/dl/go[0-9]+\.[0-9]+((\.[0-9]+)?).linux-armv6l.tar.gz")
archive_name=$(echo ${url} | cut -d '/' -f5)
wget ${url}
sudo tar -C /usr/local -xvf- ${archive_name}
cat >> ~/.bashrc << 'EOF'
export GOPATH=$HOME/go
export PATH=/usr/local/go/bin:$PATH:$GOPATH/bin
EOF
rm ${archive_name}
source ~/.bashrc


mkdir -p ~/code/go/src/github.com/johnusher
git clone https://github.com/johnusher/ardpifi.git ~/code/go/src/github.com/


export GOPATH=$HOME/code/go/src

cat >> ~/.bashrc << 'EOF'
export GOPATH=$HOME/code/go/src
export PATH=/usr/local/go/bin:$PATH:$GOPATH/bin
EOF

source ~/.bashrc
```

NB above has a tar problem for me:

curl -O https://dl.google.com/go/go1.15.6.linux-armv6l.tar.gz

tar -xvf go1.15.6.linux-armv6l.tar.gz
sudo mv go /usr/local


## Run

```bash
go run jumain.go
```

Press any key to print to screen (and eventually send to arduino).

To exit, press "q" to exit termbox, and then ctrl-c to exit the program.


# Arduino CLI install

folow instructions here:
https://siytek.com/arduino-cli-raspberry-pi/

+
arduino-cli core install arduino:avr



<del> 
## add libraries:
arduino-cli lib search Adafruit_NeoPixel
</del>

in duino_src:
compile:
arduino-cli compile --fqbn arduino:avr:uno duino_src

flash:
arduino-cli upload -p /dev/ttyACM0 --fqbn arduino:avr:uno duino_src

# Run
go run jumain.go