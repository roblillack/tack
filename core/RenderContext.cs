using System;

namespace Tack
{
	public class RenderContext
	{
		public Page Page { get; set; }

		public RenderContext (Page p)
		{
			Page = p;
		}
	}
}

