package services

// Info contains information about a service commonly found on a port
type Info struct {
	Name        string
	Description string
}

// portRange defines a range of ports that share the same service
type portRange struct {
	start int
	end   int
	info  Info
}

// ranges holds port ranges for services (checked if exact match fails)
var ranges = []portRange{
	{6463, 6472, Info{Name: "Discord", Description: "Discord RPC"}},
	{6881, 6889, Info{Name: "BitTorrent", Description: "BitTorrent P2P file sharing"}},
	{27015, 27020, Info{Name: "Steam", Description: "Steam game server"}},
	{27036, 27037, Info{Name: "Steam", Description: "Steam in-home streaming"}},
}

// ports maps port numbers to their commonly known services
// This combines IANA standard ports with well-known application ports
var ports = map[int]Info{
	// IANA Standard Ports (Well-Known Ports: 0-1023)
	20:   {Name: "FTP-DATA", Description: "File Transfer Protocol (Data)"},
	21:   {Name: "FTP", Description: "File Transfer Protocol (Control)"},
	22:   {Name: "SSH", Description: "Secure Shell"},
	23:   {Name: "Telnet", Description: "Telnet remote login"},
	25:   {Name: "SMTP", Description: "Simple Mail Transfer Protocol"},
	53:   {Name: "DNS", Description: "Domain Name System"},
	67:   {Name: "DHCP", Description: "Dynamic Host Configuration Protocol (Server)"},
	68:   {Name: "DHCP", Description: "Dynamic Host Configuration Protocol (Client)"},
	69:   {Name: "TFTP", Description: "Trivial File Transfer Protocol"},
	80:   {Name: "HTTP", Description: "Hypertext Transfer Protocol"},
	110:  {Name: "POP3", Description: "Post Office Protocol v3"},
	111:  {Name: "RPCbind", Description: "ONC RPC"},
	123:  {Name: "NTP", Description: "Network Time Protocol"},
	135:  {Name: "MS-RPC", Description: "Microsoft RPC"},
	137:  {Name: "NetBIOS-NS", Description: "NetBIOS Name Service"},
	138:  {Name: "NetBIOS-DGM", Description: "NetBIOS Datagram Service"},
	139:  {Name: "NetBIOS-SSN", Description: "NetBIOS Session Service"},
	143:  {Name: "IMAP", Description: "Internet Message Access Protocol"},
	161:  {Name: "SNMP", Description: "Simple Network Management Protocol"},
	162:  {Name: "SNMP-TRAP", Description: "SNMP Trap"},
	389:  {Name: "LDAP", Description: "Lightweight Directory Access Protocol"},
	443:  {Name: "HTTPS", Description: "HTTP over TLS/SSL"},
	445:  {Name: "SMB", Description: "Server Message Block"},
	465:  {Name: "SMTPS", Description: "SMTP over TLS/SSL"},
	514:  {Name: "Syslog", Description: "System logging protocol"},
	515:  {Name: "LPD", Description: "Line Printer Daemon"},
	587:  {Name: "SMTP-Submit", Description: "SMTP Message Submission"},
	636:  {Name: "LDAPS", Description: "LDAP over TLS/SSL"},
	993:  {Name: "IMAPS", Description: "IMAP over TLS/SSL"},
	995:  {Name: "POP3S", Description: "POP3 over TLS/SSL"},

	// IANA Registered Ports (1024-49151)
	1433:  {Name: "MS-SQL", Description: "Microsoft SQL Server"},
	1434:  {Name: "MS-SQL-Monitor", Description: "Microsoft SQL Server Monitor"},
	1521:  {Name: "Oracle", Description: "Oracle Database"},
	1723:  {Name: "PPTP", Description: "Point-to-Point Tunneling Protocol"},
	1883:  {Name: "MQTT", Description: "MQ Telemetry Transport"},
	2049:  {Name: "NFS", Description: "Network File System"},
	2375:  {Name: "Docker", Description: "Docker REST API (unencrypted)"},
	2376:  {Name: "Docker-TLS", Description: "Docker REST API (TLS)"},
	3000:  {Name: "Dev-Server", Description: "Common development server port"},
	3306:  {Name: "MySQL", Description: "MySQL Database"},
	3389:  {Name: "RDP", Description: "Remote Desktop Protocol"},
	4070:  {Name: "Spotify", Description: "Spotify local discovery"},
	4371:  {Name: "Spotify", Description: "Spotify remote control"},
	5000:  {Name: "Flask/UPnP", Description: "Flask dev server or UPnP"},
	5432:  {Name: "PostgreSQL", Description: "PostgreSQL Database"},
	5672:  {Name: "AMQP", Description: "Advanced Message Queuing Protocol"},
	5900:  {Name: "VNC", Description: "Virtual Network Computing"},
	5984:  {Name: "CouchDB", Description: "CouchDB Database"},
	6379:  {Name: "Redis", Description: "Redis Database"},
	6443:  {Name: "Kubernetes", Description: "Kubernetes API Server"},
	7000:  {Name: "Cassandra", Description: "Apache Cassandra"},
	8000:  {Name: "Alt-HTTP", Description: "Alternative HTTP port"},
	8008:  {Name: "Alt-HTTP", Description: "Alternative HTTP port"},
	8080:  {Name: "HTTP-Proxy", Description: "HTTP proxy/alternative"},
	8081:  {Name: "HTTP-Alt", Description: "Alternative HTTP port"},
	8443:  {Name: "HTTPS-Alt", Description: "Alternative HTTPS port"},
	8545:  {Name: "Ethereum", Description: "Ethereum JSON-RPC"},
	8546:  {Name: "Ethereum-WS", Description: "Ethereum WebSocket"},
	8883:  {Name: "MQTT-TLS", Description: "MQTT over TLS/SSL"},
	9000:  {Name: "SonarQube", Description: "SonarQube Server"},
	9090:  {Name: "Prometheus", Description: "Prometheus metrics"},
	9092:  {Name: "Kafka", Description: "Apache Kafka"},
	9200:  {Name: "Elasticsearch", Description: "Elasticsearch HTTP"},
	9300:  {Name: "Elasticsearch", Description: "Elasticsearch Transport"},
	11211: {Name: "Memcached", Description: "Distributed memory caching"},
	15672: {Name: "RabbitMQ", Description: "RabbitMQ Management"},
	27017: {Name: "MongoDB", Description: "MongoDB Database"},
	50000: {Name: "SAP", Description: "SAP HTTP"},
	50070: {Name: "Hadoop", Description: "Hadoop NameNode"},
	57621: {Name: "Spotify", Description: "Spotify Connect"},

	// Gaming and P2P
	3074:  {Name: "Xbox-Live", Description: "Xbox Live"},
	3478:  {Name: "STUN", Description: "Session Traversal Utilities for NAT"},
	3479:  {Name: "TURN", Description: "Traversal Using Relays around NAT"},
	4000:  {Name: "Diablo-II", Description: "Diablo II game"},
	5353:  {Name: "mDNS", Description: "Multicast DNS"},
	25565: {Name: "Minecraft", Description: "Minecraft game server"},
	25575: {Name: "Minecraft-RCON", Description: "Minecraft RCON"},

	// IoT and Smart Home
	1900: {Name: "UPnP", Description: "Universal Plug and Play"},
	5683: {Name: "CoAP", Description: "Constrained Application Protocol"},
	8123: {Name: "Home-Assistant", Description: "Home Assistant"},
}
