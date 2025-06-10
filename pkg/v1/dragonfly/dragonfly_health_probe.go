package dragonfly

const healthProbe = `
#!/bin/sh

echo PING | openssl s_client -connect 127.0.0.1:6379 -CAfile /etc/dragonfly/tls/ca.crt -quiet -no_ign_eof

exit $?
`
