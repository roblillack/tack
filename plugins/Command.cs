using System;
using System.Collections.Generic;

namespace Tack.Plugins
{
	public abstract class Command
	{
		public abstract string Name { get; }
		public abstract string Description { get; }
		public abstract void Execute(IList<string> parameters);

		public virtual bool IsVisible { get { return true; } }
		public Tack Application { set; get; }

		public class Error : Exception
		{
			public Error (string msg) : base (msg) {}
		}
	}
}

