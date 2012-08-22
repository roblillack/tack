using System;
using System.Collections.Generic;

namespace Tack
{
	public static class DictUtils
	{
		public static void AddAll<TKey, TValue> (this IDictionary<TKey, TValue> me, IDictionary<TKey, TValue> other)
		{
			foreach (var i in other) {
				me [i.Key] = i.Value;
			}
		}

		public static IDictionary<TKey, TValue> Combine<TKey, TValue> (params IDictionary<TKey, TValue> [] dicts)
		{
			var r = new Dictionary<TKey, TValue> ();
			foreach (var i in dicts) {
				r.AddAll (i);
			}
			return r;
		}
	}
}

