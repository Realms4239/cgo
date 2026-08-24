export function Skeleton({ lines = 3 }: { lines?: number }) {
  return (
    <div className="c-skel" aria-hidden>
      {Array.from({ length: lines }, (_, i) => <div key={i} className="c-skel-line" />)}
    </div>
  );
}
