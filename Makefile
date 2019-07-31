PREFIX=/usr/local
VERSION=`git describe`

all: build-mac build-linux

uninstall:
	rm -rf ${PREFIX}/bin/tack ${PREFIX}/lib/tack ${PREFIX}/share/tack

run:
	dotnet run tack.csproj

build-mac:
	dotnet publish -c Release -r osx-x64

install-mac: build-mac
	install -d ${PREFIX}/lib/tack
	install bin/Release/netcoreapp2.2/osx-x64/publish/* ${PREFIX}/lib/tack
	install tack.sh ${PREFIX}/bin/tack
	chmod a+rx ${PREFIX}/bin/tack

package-mac: build-mac
	rm -rf tmp
	install -d tmp/lib/tack tmp/bin
	install bin/Release/netcoreapp2.2/osx-x64/publish/* tmp/lib/tack
	install tack.sh tmp/bin/tack
	chmod a+rx tmp/bin/tack
	tar -C tmp -czvf tack_${VERSION}_osx-x64.tgz bin lib
	rm -rf tmp

build-linux:
	dotnet publish -c Release -r linux-x64

install-linux: build-linux
	install -d ${PREFIX}/lib/tack
	install bin/Release/netcoreapp2.2/linux-x64/publish/* ${PREFIX}/lib/tack
	install tack.sh ${PREFIX}/bin/tack
	chmod a+rx ${PREFIX}/bin/tack

package-linux: build-linux
	rm -rf tmp
	install -d tmp/lib/tack tmp/bin
	install bin/Release/netcoreapp2.2/linux-x64/publish/* tmp/lib/tack
	install tack.sh tmp/bin/tack
	chmod a+rx tmp/bin/tack
	tar -C tmp -czvf tack_${VERSION}_linux-x64.tgz bin lib
	rm -rf tmp

clean:
	rm -rf bin/ obj/

.PHONY: uninstall build-mac install-mac run build-linux install-linux package-linux package-mac clean
