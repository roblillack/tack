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

		string template;
		IDictionary<string, object> variables;

		public Page (Tacker tacker, string realpath)
		{
			Tacker = tacker;
			DiskPath = realpath;
			Permalink = realpath.Replace (Tacker.ContentDir, "");
			Name = Path.GetFileName (realpath);
		}

		public string Template {
			get {
				if (template != null) {
					return template;
				}
				LoadVariables ();
				if (template == null) {
					throw new FileNotFoundException ("No Template found for page " + Permalink);
				}
				return template;
			}
		}

		public IDictionary<string, object> Variables {
			get {
				return variables ?? (variables = LoadVariables ());
			}
		}

		private IDictionary<string, object> LoadVariables ()
		{
			var metadata = new Dictionary<string, object> ();
			var assets = new HashSet<string> ();

			foreach (var i in Directory.GetFiles (DiskPath, "*")) {
				var map = Tacker.ProcessMetadata (i);
				if (map != null) {
					template = template ?? Path.GetFileNameWithoutExtension (i);
					metadata.AddAll (map);
					continue;
				}

				assets.Add (i.Replace (DiskPath, ""));
			}

			this.Assets = assets;

			return metadata;
		}

		public void Generate ()
		{
			Directory.CreateDirectory (Tacker.TargetDir + Permalink);

			var data = DictUtils.Combine (Tacker.Metadata, Variables);
			using (var writer = File.CreateText(Path.Combine (Tacker.TargetDir + Permalink, "index.html"))) {
				Tacker.FindTemplate (Template).Render (data, writer, Tacker.FindTemplate);
			}

			/*Nustache.Core.Render.FileToFile (Path.Combine (Tacker.TemplateDir, Template + ".mustache"),
			                                 ,
			                                 Path.Combine (Tacker.TargetDir + Permalink, "index.html"));*/
		}
	}
}

