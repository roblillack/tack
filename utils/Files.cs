using System;
using System.Collections.Generic;
using System.IO;

namespace Tack.Utils
{
	public class Files
	{
		public static IEnumerable<string> EnumerateAllDirs (string path, bool includeTopLevel = true)
		{
			if (includeTopLevel && Directory.Exists (path)) {
				yield return path;
			}
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

		public static IEnumerable<string> EnumerateAllSubdirs (string path)
		{
			return EnumerateAllDirs (path, false);
		}

		public static IEnumerable<string> EnumerateAllFiles (string path)
		{
			foreach (var dir in Files.EnumerateAllDirs (path)) {
				foreach (var i in Files.GetAllFiles (dir)) {
					yield return i;
				}
			}
		}

		public static IEnumerable<string> FindDirsWithFiles(string path, ICollection<string> extensions)
		{
			// FIXME: There seems to be a bug in Mono's Directory.EnumerateFiles implementation
			foreach (var dir in EnumerateAllSubdirs (path)) {
				var found = false;
				foreach (var i in GetAllFiles (dir)) {
					if (extensions.Contains (Path.GetExtension (i).Replace (".", ""))) {
						found = true;
					}
				}
				if (found) {
					yield return dir;
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

