

```shell
cat <<EOF > /etc/systemd/system/iptables-restore.service
[Unit]
Description=Restore iptables
After=network.target

[Service]
Type=oneshot
ExecStart=/sbin/iptables-restore /etc/sysconfig/iptables

[Install]
WantedBy=multi-user.target
EOF

```

```shell
iptables-save > /etc/sysconfig/iptables
systemctl daemon-reexec
systemctl enable iptables-restore
```