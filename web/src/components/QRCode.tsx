import { QRCodeSVG } from 'qrcode.react';

interface Props {
  value: string;
  size?: number;
}

export function QRCode({ value, size = 200 }: Props) {
  return (
    <div className="bg-white p-4 rounded-lg inline-block">
      <QRCodeSVG value={value} size={size} level="M" />
    </div>
  );
}
