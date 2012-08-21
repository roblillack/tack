using System;
using System.Collections.Generic;
using System.IO;
using Tack.Plugins;

namespace Tack.Plugins.Base
{
	public class Tack : Command
	{
		public override string Name {
			get { return "tack"; }
		}
		
		public override string Description {
			get { return "Tacks up everything"; }
		}
		
		public override void Execute (IList<string> parameters)
		{
			var tacker = new Tacker (Directory.GetCurrentDirectory ());
			tacker.Logger = (x,y) => Console.WriteLine (x, y);
			tacker.Tack ();
		}
	}
}

