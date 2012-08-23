using System;
using System.Collections.Generic;
using System.IO;
using YamlDotNet.RepresentationModel;
using Nustache.Core;
using MarkdownSharp;

namespace Tack
{
	public class Tacker
	{
		static readonly string[] TEMPLATE_LANGS = { "mustache" };
		static readonly string[] METADATA_LANGS = { "yml" };
		static readonly string[] MARKUP_LANGS = { "mkd" };

		public delegate void LogFn (string format, params object[] args);

		public string BaseDir { get; protected set; }
		public string ContentDir { get { return Path.Combine (BaseDir, "content"); } }
		public string TemplateDir { get { return Path.Combine (BaseDir, "templates"); } }
		public string TargetDir { get { return Path.Combine (BaseDir, "output"); } }
		public string AssetDir { get { return Path.Combine (BaseDir, "public"); } }
		public IDictionary<string, object> Metadata { get; protected set; }
		public LogFn Logger { get; set; }
		public ISet<Page> Pages { get; protected set; }
		public IList<Page> Navigation { get; protected set; }

		Markdown markdown;

		public Tacker (string dir)
		{
			markdown = new Markdown ();
			markdown.AutoHyperlink = true;

			BaseDir = dir;
			Metadata = LoadMetadata ();
			Pages = FindAllPages ();

			foreach (var i in Pages) {
				i.Init ();
			}

			var navi = new SortedDictionary<string, Page> ();
			foreach (var i in Pages) {
				if (i.Parent == null && !i.IsFloating) {
					navi.Add (Path.GetFileName (i.DiskPath), i);
				}
			}
			Navigation = new List<Page> (navi.Values);
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
			Log ("Tacking up {0} ({1} pages)", BaseDir, Pages.Count);

			foreach (var page in Pages) {
				Log ("{0} => {1} (template: {2})", page.Permalink, page.Name, page.Template);
				page.Generate ();
			}

			foreach (var i in FindAllAssets ()) {
				var dest = i.Replace (AssetDir, TargetDir);
				Directory.CreateDirectory (Path.GetDirectoryName (dest));
				File.Copy (i, dest, true);
				Console.WriteLine ("Copying {0}", i);
			}
		}

		public Template FindTemplate (string name)
		{
			foreach (var ext in TEMPLATE_LANGS) {
				var tpl = Path.Combine (TemplateDir, name.Trim () + "." + ext);

				if (!File.Exists (tpl)) {
					continue;
				}

				using (var reader = File.OpenText (tpl)) {
					var template = new Template ();
					template.Load (reader);
					return template;
				}
			}

            return null;
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
            foreach (var dir in Directory.EnumerateDirectories (path, "*", SearchOption.AllDirectories)) {
				foreach (var i in GetAllFiles (dir)) {
					foreach (var ext in extensions) {
						if (i.EndsWith ("." + ext)) {
							yield return dir;
						}
					}
				}
			}
		}

		string[] GetAllFiles (string dir)
		{
			try {
				return Directory.GetFiles (dir, "*", SearchOption.TopDirectoryOnly);
			} catch (UnauthorizedAccessException) {
				return new string[]{};
			}
		}

		IEnumerable<string> FindAllAssets ()
		{
			// FIXME: There seems to be a bug in Mono's Directory.EnumerateFiles implementation
			foreach (var i in GetAllFiles (AssetDir)) {
				yield return i;
			}

			foreach (var dir in Directory.EnumerateDirectories (AssetDir, "*", SearchOption.AllDirectories)) {
				foreach (var i in GetAllFiles (dir)) {
					yield return i;
				}
			}
		}

		public IDictionary<string, object> ProcessMetadata (string file)
		{
			foreach (var ext in METADATA_LANGS) {
				if (file.EndsWith ("." + ext)) {
					var map = new Dictionary<string, object> ();
					var stream = new YamlStream ();
					stream.Load (new StreamReader (file));

					foreach (var doc in stream.Documents) {
						if (doc.RootNode is YamlMappingNode) {
							var seq = doc.RootNode as YamlMappingNode;
							foreach (var node in seq.Children) {
								var key = node.Key as YamlScalarNode;
								object val = node.Value;
								if (val is YamlScalarNode && (val as YamlScalarNode).Style == YamlDotNet.Core.ScalarStyle.Literal) {
									Console.WriteLine ("{0} --> {1}", key, (val as YamlScalarNode).Style);
									val = markdown.Transform (val.ToString ());
								}
								map.Add (key.Style == YamlDotNet.Core.ScalarStyle.Plain ?
								         key.Value.Substring (1) : key.Value,
								         val);
							}
						}
					}
					return map;
				}
			}

			// Not a known meta-data format
			return null;
		}

		private IDictionary<string, object> LoadMetadata ()
		{
			var metadata = new Dictionary<string, object> ();
			foreach (var file in GetAllFiles (BaseDir)) {
				var map = ProcessMetadata (file);
				if (map != null) {
					metadata.AddAll (map);
				}
			}
			return metadata;
		}
	}
}

