using System;
using System.Collections.Generic;
using System.IO;

namespace Tack
{
	public class Tacker
	{
		private static readonly string[] TEMPLATE_LANGS = { "mustache" };
		private static readonly string[] METADATA_LANGS = { "yml" };
		public delegate void LogFn (string format, params object[] args);

		public string BaseDir { get; protected set; }
		public LogFn Logger { get; set; } 

		public Tacker (string dir)
		{
			BaseDir = dir;
		}

		protected void Log (string format, params object[] args)
		{
			if (Logger == null) {
				Console.WriteLine (format, args);
			} else {
				Logger (format, args);
			}
		}

		public void Tack ()
		{
			var pages = FindAllPages ();
			Log ("Tacking up {0}", BaseDir);
			Log ("{0} Templates found.", FindAllTemplates ().Count);
			Log ("{0} Pages found.", pages.Count);

			foreach (var page in pages) {
				foreach (var entry in page.GetVariables ()) {
					Log ("{0} => {1}", entry.Key, entry.Value);
				}
			}
		}

		ISet<string> FindAllTemplates()
		{
			var set = new HashSet<string> ();
			foreach (var i in Directory.EnumerateFiles (Path.Combine (BaseDir, "templates"),
			                                            "*",
			                                            SearchOption.AllDirectories)) {
				foreach (var extension in TEMPLATE_LANGS) {
					if (i.EndsWith ("." + extension)) {
						set.Add (i);
					}
				}
			}
			return set;
		}

		ISet<Page> FindAllPages()
		{
			var set = new HashSet<Page> ();
			foreach (var i in Directory.EnumerateFiles (Path.Combine (BaseDir, "content"),
			                                            "*",
			                                            SearchOption.AllDirectories)) {
				foreach (var extension in METADATA_LANGS) {
					if (i.EndsWith ("." + extension)) {
						set.Add (new Page (this, i));
					}
				}
			}
			return set;
		}
	}
}

