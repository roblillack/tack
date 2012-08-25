PREFIX=/usr/local

all: bin/Release/tack.exe

bin/Release/tack.exe: lib/markdownsharp/bin/Release/MarkdownSharp.dll ext/nustache/Nustache.Core/bin/Release/Nustache.Core.dll
	xbuild /property:Configuration=Release tack.csproj

lib/markdownsharp/bin/Release/MarkdownSharp.dll:
	xbuild /property:Configuration=Release lib/markdownsharp/MarkdownSharp.csproj

ext/nustache/Nustache.Core/Nustache.Core.csproj:
	git submodule init
	git submodule update

ext/nustache/Nustache.Core/bin/Release/Nustache.Core.dll: ext/nustache/Nustache.Core/Nustache.Core.csproj
	xbuild /property:Configuration=Release ext/nustache/Nustache.Core/Nustache.Core.csproj

install: bin/Release/tack.exe
	install -d ${PREFIX}/share/tack
	install bin/Release/* ${PREFIX}/share/tack
	printf '#!/bin/sh\nmono '${PREFIX}'/share/tack/tack.exe $$@\n' > ${PREFIX}/bin/tack
	chmod a+rx ${PREFIX}/bin/tack

uninstall:
	rm -rf ${PREFIX}/bin/tack ${PREFIX}/share/tack

clean:
	rm -rf bin/ obj/ ext/nustache/bin/ ext/nustache/obj/ lib/markdownsharp/obj/ lib/markdownsharp/bin/

.PHONY: uninstall
