
# Define versions
VERSION_MAJOR?=0
VERSION_MINOR?=0
VERSION_MAINT?=0

# Jenkins run-time build id
VERSION_BUILD?=0

PKG_VERSION=$(VERSION_MAJOR).$(VERSION_MINOR).$(VERSION_MAINT)-$(VERSION_BUILD)

PKG_NAME=sdLite
PKG_NAME_FULL=$(PKG_NAME)_$(PKG_VERSION)

STATIC_FILES = $(shell find static/ -follow -type f)

all: gopacker

gopacker: main.go gopack.go
	go build

gopack.go: gopack.pl $(STATIC_FILES)
	./gopack.pl static/

sdLite-deb: gopacker
	@echo "===== Seed Drive UI ====="
	mkdir -p tmp/$(PKG_NAME_FULL)
	mkdir -p tmp/$(PKG_NAME_FULL)/opt/r1soft/$(PKG_NAME)/bin/
	cp $(PKG_NAME) tmp/$(PKG_NAME_FULL)/opt/r1soft/$(PKG_NAME)/bin/
	mkdir -p tmp/$(PKG_NAME_FULL)/etc/init/
	cp files/upstart-sdLite.conf tmp/$(PKG_NAME_FULL)/etc/init/$(PKG_NAME).conf
	mkdir -p tmp/$(PKG_NAME_FULL)/DEBIAN 
	echo "Package: $(PKG_NAME)" > tmp/$(PKG_NAME_FULL)/DEBIAN/control
	echo "Conflicts: " >> tmp/$(PKG_NAME_FULL)/DEBIAN/control   
	echo "Version: $(PKG_VERSION)" >> tmp/$(PKG_NAME_FULL)/DEBIAN/control 
	echo "Section: base" >> tmp/$(PKG_NAME_FULL)/DEBIAN/control 
	echo "Priority: required" >> tmp/$(PKG_NAME_FULL)/DEBIAN/control
	echo "Architecture: amd64" >> tmp/$(PKG_NAME_FULL)/DEBIAN/control
	echo "Depends: " >> tmp/$(PKG_NAME_FULL)/DEBIAN/control
	echo "Maintainer: R1Soft Development  <Development@r1soft.com>" >> tmp/$(PKG_NAME_FULL)/DEBIAN/control
	echo "Description: R1Soft's $(PKG_NAME) service" >> tmp/$(PKG_NAME_FULL)/DEBIAN/control
	cp files/sdLite.deb.postinst tmp/$(PKG_NAME_FULL)/DEBIAN/postinst
	chmod 755 tmp/$(PKG_NAME_FULL)/DEBIAN/postinst
	chmod 755 tmp/$(PKG_NAME_FULL)/DEBIAN/preinst
	cd tmp && fakeroot dpkg-deb --build $(PKG_NAME_FULL)
	mkdir -p targets
	mv tmp/$(PKG_NAME_FULL).deb targets/

clean:
	-@rm gopacker gopack.go 2>/dev/null || true
