using System;
using System.Collections.Generic;

namespace Tack.Utils
{
	public static class Collections
	{
		public static void AddAll<TKey, TValue> (this IDictionary<TKey, TValue> me, IDictionary<TKey, TValue> other)
		{
			foreach (var i in other) {
				me [i.Key] = i.Value;
			}
		}

		public static void AddAll<T> (this ICollection<T> me, ICollection<T> other)
		{
			foreach (var i in other) {
				me.Add (i);
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

		public static T Combine<T, X> (params T[] collections) where T : class, ICollection<X>, new()
		{
			var r = new T ();
			foreach (var i in collections) {
				r.AddAll (i);
			}
			return r;
		}

		public static ISet<X> CombinedSet<T, X> (params T[] collections) where T : class, ICollection<X>, new()
		{
			var r = new HashSet<X> ();
			foreach (var i in collections) {
				r.AddAll (i);
			}
			return r;
		}

		public static ISet<X> CombinedSet<X> (params X[][] arrays)
		{
			var r = new HashSet<X> ();
			foreach (var i in arrays) {
				r.AddAll (i);
			}
			return r;
		}
	}
}

