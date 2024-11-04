#!/bin/sh
for u in $(w -h | awk '{if (!seen[$1]++) print $1}'); do
	systemctl -M "$u@" --user daemon-reload
done
udevadm control --reload