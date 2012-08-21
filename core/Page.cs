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
		public string Template { get; protected set; }
		public Tacker Tacker { get; protected set; }

		public Page (Tacker tacker, string path)
		{
			Tacker = tacker;
			DiskPath = path;
			Permalink = path.Replace (tacker.BaseDir, "");
			Name = Path.GetFileNameWithoutExtension (path);
		}

		public IDictionary<string, object> GetVariables ()
		{
			var map = new Dictionary<string, object> ();

			var stream = new YamlStream ();
			stream.Load (new StreamReader (DiskPath));

			foreach (var doc in stream.Documents) {
				if (doc.RootNode is YamlMappingNode) {
					var seq = doc.RootNode as YamlMappingNode;
					foreach (var node in seq.Children) {
						map.Add (node.Key.ToString (), node.Value);
					}
				}
			}

			return map;
		}
	}
}

