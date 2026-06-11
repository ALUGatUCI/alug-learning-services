#!/bin/bash

# Create student user from env vars passed by Go
if [ -n "$STUDENT_USER" ]; then
    useradd -m -s /bin/bash "$STUDENT_USER" 2>/dev/null || true
    echo "$STUDENT_USER:${STUDENT_PASS:-student}" | chpasswd
    usermod -aG sudo "$STUDENT_USER"
    echo "$STUDENT_USER ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers
fi

exec /sbin/init