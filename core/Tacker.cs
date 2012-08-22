using System;
using System.Collections.Generic;
using System.IO;

namespace Tack
{
	public class Tacker
	{
		const string CONTENT_DIRECTORY = "content";
		const string TEMPLATE_DIRECTORY = "templates";
		static readonly string[] TEMPLATE_LANGS = { "mustache" };
		static readonly string[] METADATA_LANGS = { "yml" };

		public delegate void LogFn (string format, params object[] args);

		public string BaseDir { get; protected set; }
		public string ContentDir { get { return Path.Combine (BaseDir, CONTENT_DIRECTORY); } }
		public string TemplateDir { get { return Path.Combine (BaseDir, TEMPLATE_DIRECTORY); } }
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
				Log ("{0} => {1} (template: {2})", page.Permalink, page.Name, page.Template);
				Log ("  {0}", String.Join (", ", page.Variables.Keys));
				//foreach (var entry in page.GetVariables ()) {
					//Log ("{0} => {1}", entry.Key, entry.Value);
				//}
			}
		}

		ISet<string> FindAllTemplates()
		{
			var set = new HashSet<string> ();
			foreach (var i in Directory.EnumerateFiles (TemplateDir, "*",
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
			foreach (var i in FindDirsWithFiles (ContentDir, METADATA_LANGS)) {
				set.Add (new Page (this, i));
			}
			return set;
		}

		IEnumerable<string> FindDirsWithFiles(string path, params string[] extensions)
		{
            foreach (var dir in Directory.EnumerateDirectories (path, "*", SearchOption.AllDirectories))
            {
				// FIXME: There seems to be a bug in Mono's Directory.EnumerateFiles implementation
                string[] files;
                try {
                    files = Directory.GetFiles (dir, "*");
                } catch (UnauthorizedAccessException) {
                    continue;
                }
				foreach (var i in files) {
					foreach (var ext in extensions) {
						if (i.EndsWith ("." + ext)) {
							yield return dir;
						}
					}
				}
			}
		}
	}
}

