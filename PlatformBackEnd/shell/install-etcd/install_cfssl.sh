#/bin/bash


echo "--------insdtall cfssl------------"
yum install -y wget

cp -f cfssl /usr/local/bin/cfssl
cp -f cfssljson /usr/local/bin/cfssljson
cp -f cfssl-certinfo /usr/local/bin/cfssl-certinfo
chmod +x /usr/local/bin/cfssl /usr/local/bin/cfssljson /usr/local/bin/cfssl-certinfo