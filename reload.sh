#!/bin/sh
for U in $(w -h | awk '{print $1}' | sort -u); do
	systemctl -M "$U@" --user daemon-reload
done
udevadm control --reload