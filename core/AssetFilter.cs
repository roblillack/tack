using System;

namespace Tack
{
	public interface AssetFilter
	{
		string[] Extensions { get; }
		void Filter (Tacker tacker, string srcfile);
	}
}

