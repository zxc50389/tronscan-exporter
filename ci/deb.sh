#!/bin/sh

cd "$(dirname "$0")"

sed -i "3iSERVICENAME=${SERVICENAME}" deb/DEBIAN/preinst
sed -i "3iSERVICENAME=${SERVICENAME}" deb/DEBIAN/postinst
sed -i "3iSERVICENAME=${SERVICENAME}" deb/DEBIAN/prerm
sed -i "1iPackage: ${SERVICENAME}" deb/DEBIAN/control
echo "Description: ${SERVICENAME}" >> deb/DEBIAN/control
echo "Version: ${VERSION}" >> deb/DEBIAN/control
cat ./deb/DEBIAN/control

mkdir -p ./deb/etc/systemd/system
cat << EOF >>./deb/etc/systemd/system/${SERVICENAME}.service
[Unit]
Description=${SERVICENAME}
After=network-online.target

[Service]
WorkingDirectory=/becreator/${SERVICENAME}
ExecStart=/becreator/${SERVICENAME}/${SERVICENAME}
ExecStop=/usr/bin/pkill ${SERVICENAME}

[Install]
WantedBy=multi-user.target
EOF

chmod 755 deb/DEBIAN/preinst
chmod 755 deb/DEBIAN/postinst
chmod 755 deb/DEBIAN/prerm
chmod 755 deb/DEBIAN

dpkg -b deb ${SERVICENAME}.deb
ls -al
