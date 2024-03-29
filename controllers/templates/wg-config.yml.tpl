[Interface]
Address ={{.Address}}
ListenPort = 51820
PrivateKey = {{.PrivateKey}}

PostUp = iptables -A FORWARD -i %i -j ACCEPT;
PostUp = iptables -A FORWARD -o %i -j ACCEPT; 
PostUp = iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE;

PostUp = iptables -t nat -A PREROUTING -i eth0 -p udp --dport 51820 -j ACCEPT;

PostDown = iptables -D FORWARD -i %i -j ACCEPT;
PostDown = iptables -D FORWARD -o %i -j ACCEPT; 
PostDown = iptables -t nat -D POSTROUTING -o eth0 -j MASQUERADE;

PostDown = iptables -t nat -D PREROUTING -i eth0 -p udp --dport 51820 -j ACCEPT;

{{- range $value := .Peers }}
[Peer]
PublicKey = {{ $value.PublicKey }}
AllowedIPs = {{ $value.AllowedIps }}
{{- end}}
