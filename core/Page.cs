using System;
using System.Collections.Generic;
using System.IO;
using YamlDotNet.RepresentationModel;

namespace Tack
{
	public class Page
	{
		public string Name { get; protected set; }
		public string DiskPath { get; protected set; }
		public string Permalink { get; protected set; }
		public Tacker Tacker { get; protected set; }
		public ISet<string> Assets { get; protected set; }
		public IDictionary<string, object> Variables { get; protected set; }
		public string Template { get; protected set; }

		public Page (Tacker tacker, string realpath)
		{
			Tacker = tacker;
			DiskPath = realpath;
			Permalink = realpath.Replace (Tacker.ContentDir, "");
			Name = Path.GetFileName (realpath);

			Init ();

			if (Template == null) {
				throw new FileNotFoundException ("No Template found for page " + Permalink);
			}
		}

		private void Init ()
		{
			var metadata = new Dictionary<string, object> ();
			var assets = new HashSet<string> ();

			foreach (var i in Directory.GetFiles (DiskPath, "*")) {
				var map = Tacker.ProcessMetadata (i);
				if (map != null) {
					Template = Template ?? Path.GetFileNameWithoutExtension (i);
					metadata.AddAll (map);
					continue;
				}

				assets.Add (i.Replace (DiskPath, ""));
			}

			Assets = assets;
			Variables = metadata;
		}

		public void Generate ()
		{
			Directory.CreateDirectory (Tacker.TargetDir + Permalink);

			var data = DictUtils.Combine (Tacker.Metadata, Variables);
			using (var writer = File.CreateText(Path.Combine (Tacker.TargetDir + Permalink, "index.html"))) {
				Tacker.FindTemplate (Template).Render (data, writer, Tacker.FindTemplate);
			}
		}
	}
}

