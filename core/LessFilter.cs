using System;
using System.IO;
using dotless.Core;

namespace Tack
{
	public class LessFilter : AssetFilter
	{
		public string[] Extensions {
			get {
				return new string[] { "less" };
			}
		}

		public void Filter (Tacker tacker, string src)
		{
			var dir = Path.GetDirectoryName (src.Replace (tacker.AssetDir, tacker.TargetDir));
			var dest = Path.Combine (dir, Path.GetFileNameWithoutExtension (src) + ".css");
			File.WriteAllText (dest, Less.Parse (File.ReadAllText (src)));
		}
	}
}

