import type { ToastsContextType } from '../components/toasts';

export async function copyWalletAddress(
  address: string | undefined,
  copyToClipboard: (text: string) => Promise<void> | void,
  toast: ToastsContextType['toast'],
): Promise<boolean> {
  if (!address) return false;

  try {
    await copyToClipboard(address);
    toast('success', 'Address copied to clipboard', { autoClose: 1500 });
    return true;
  } catch {
    toast('error', 'Could not copy your address. Please try again.');
    return false;
  }
}
