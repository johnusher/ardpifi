#!/bin/sh


# script to run batman mesh/

# Based on:
# https://www.reddit.com/r/darknetplan/comments/68s6jp/how_to_configure_batmanadv_on_the_raspberry_pi_3/

# sudo apt install libnl-3-dev libnl-genl-3-dev
# git clone https://git.open-mesh.org/batctl.git
# (
#   cd batctl
#   sudo make install
# )


# Activate batman-adv
sudo modprobe batman-adv
# Disable and configure wlan0
sudo ip link set wlan0 down
# sudo ifconfig wlan0 mtu 1532
#sudo killall wpa_supplicant # ???
wpa_cli terminate -i wlan0 # this works 
sudo iwconfig wlan0 mode ad-hoc
sudo iwconfig wlan0 essid ledmesh
sudo iwconfig wlan0 ap any
sudo iwconfig wlan0 channel 8
sleep 1s
sudo ip link set wlan0 up
sleep 1s
sudo batctl if add wlan0
sleep 1s
sudo ifconfig bat0 up
sleep 5s
# Use different IPv4 addresses for each device
# This is the only change necessary to the script for
# different devices. Make sure to indicate the number
# of bits used for the mask.
sudo ifconfig bat0 172.27.0.1/16   # 172.27.0.x
sudo iwconfig wlan0 ap CA:B4:54:B1:5A:75

# test batctl
sudo batctl o