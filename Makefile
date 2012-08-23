all: nustache markdownsharp

markdownsharp:
	xbuild /property:Configuration=Release lib/markdownsharp/MarkdownSharp.csproj

nustache:
	xbuild /property:Configuration=Release ext/nustache/Nustache.Core/Nustache.Core.csproj
