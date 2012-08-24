using System;
using System.Collections.Generic;
using System.IO;

namespace Tack.Utils
{
	public class Files
	{
		public static IEnumerable<string> EnumerateAllSubdirs (string path)
		{
			var e = Directory.EnumerateDirectories (path, "*", SearchOption.AllDirectories).GetEnumerator ();
			for (;;) {
				try {
					if (!e.MoveNext ()) {
						yield break;
					}
				} catch (DirectoryNotFoundException) {
					yield break;
				}
				yield return e.Current;
			}
		}

		public static IEnumerable<string> EnumerateAllFiles (string path)
		{
			foreach (var i in Files.GetAllFiles (path)) {
				yield return i;
			}

			foreach (var dir in Files.EnumerateAllSubdirs (path)) {
				foreach (var i in Files.GetAllFiles (dir)) {
					yield return i;
				}
			}
		}

		public static IEnumerable<string> FindDirsWithFiles(string path, ICollection<string> extensions)
		{
			// FIXME: There seems to be a bug in Mono's Directory.EnumerateFiles implementation
			foreach (var dir in EnumerateAllSubdirs (path)) {
				foreach (var i in GetAllFiles (dir)) {
					if (extensions == null) {
						yield return dir;
					}
					foreach (var ext in extensions) {
						if (i.EndsWith ("." + ext)) {
							yield return dir;
						}
					}
				}
			}
		}

		public static string[] GetAllFiles (string dir)
		{
			try {
				return Directory.GetFiles (dir, "*", SearchOption.TopDirectoryOnly);
			} catch (Exception) {
				return new string[]{};
			}
		}

	}
}

