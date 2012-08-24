using System;
using System.Collections.Generic;
using System.IO;
using YamlDotNet.RepresentationModel;
using Nustache.Core;
using MarkdownSharp;
using Tack.Utils;

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
		IDictionary<string, AssetFilter> assetFilters;

		public Tacker (string dir)
		{
			markdown = new Markdown ();
			markdown.AutoHyperlink = true;

			assetFilters = new Dictionary<string, AssetFilter> ();
			assetFilters.Add ("less", new LessFilter ());

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

			Directory.Delete (TargetDir, true);

			foreach (var page in Pages) {
				Log ("{0} => {1} (template: {2})", page.Permalink, page.Name, page.Template);
				page.Generate ();
			}

			foreach (var i in FindAllAssets ()) {
				var dest = i.Replace (AssetDir, TargetDir);
				Directory.CreateDirectory (Path.GetDirectoryName (dest));
				if (assetFilters.ContainsKey (Path.GetExtension (i).Replace (".", ""))) {
					var filter = assetFilters [Path.GetExtension (i).Replace (".", "")];
					Log ("Applying {0} to {1} ...", filter.GetType ().Name, i);
					filter.Filter (this, i);
				} else {
					Console.WriteLine ("Copying {0}", i);
					File.Copy (i, dest, true);
				}
			}
		}

		public Template FindTemplate (string name)
		{
			if (name == null) {
				name = "default";
			}

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
			foreach (var i in Files.FindDirsWithFiles (ContentDir, Collections.CombinedSet (MARKUP_LANGS, METADATA_LANGS))) {
				set.Add (new Page (this, i));
			}
			return set;
		}

		IEnumerable<string> FindAllAssets ()
		{
			return Files.EnumerateAllFiles (AssetDir);
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

		public string ProcessMarkup (string file)
		{
			foreach (var ext in MARKUP_LANGS) {
				if (Path.GetExtension (file).Equals ("." + ext)) {
					return markdown.Transform (File.ReadAllText (file));
				}
			}

			return null;
		}

		private IDictionary<string, object> LoadMetadata ()
		{
			var metadata = new Dictionary<string, object> ();
			foreach (var file in Files.GetAllFiles (BaseDir)) {
				var map = ProcessMetadata (file);
				if (map != null) {
					metadata.AddAll (map);
				}
			}
			return metadata;
		}
	}
}

