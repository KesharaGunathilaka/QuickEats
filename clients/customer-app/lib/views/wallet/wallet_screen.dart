import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:quickeats_customer/core/constants/app_colors.dart';
import 'package:quickeats_customer/viewmodels/wallet_viewmodel.dart';

class WalletScreen extends StatefulWidget {
  const WalletScreen({super.key});
  @override State<WalletScreen> createState() => _WalletScreenState();
}

class _WalletScreenState extends State<WalletScreen> {
  @override void initState() { super.initState(); context.read<WalletViewModel>().loadWallet(); }

  @override Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Wallet', style: TextStyle(fontWeight: FontWeight.w700)), backgroundColor: AppColors.white, elevation: 0),
      body: Consumer<WalletViewModel>(builder: (_, vm, __) {
        if (vm.isLoading) return const Center(child: CircularProgressIndicator());
        return Padding(padding: const EdgeInsets.all(16), child: Column(crossAxisAlignment: CrossAxisAlignment.stretch, children: [
          // Balance card
          Container(padding: const EdgeInsets.all(28), decoration: BoxDecoration(gradient: const LinearGradient(colors: [AppColors.primary, const Color(0xFF1D4ED8)], begin: Alignment.topLeft, end: Alignment.bottomRight), borderRadius: BorderRadius.circular(20)),
            child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
              const Text('Available Balance', style: TextStyle(color: Colors.white70, fontSize: 14)),
              const SizedBox(height: 8),
              Text(vm.wallet?.formattedBalance ?? 'LKR 0.00', style: const TextStyle(color: Colors.white, fontSize: 36, fontWeight: FontWeight.w800)),
              const SizedBox(height: 4),
              const Text('QuickEats Wallet', style: TextStyle(color: Colors.white60, fontSize: 13)),
            ]),
          ),
          const SizedBox(height: 24),

          const Text('Quick Top-Up', style: TextStyle(fontSize: 16, fontWeight: FontWeight.w600)),
          const SizedBox(height: 12),
          Wrap(spacing: 10, runSpacing: 10, children: [200, 500, 1000, 2000].map((amt) => SizedBox(width: (MediaQuery.of(context).size.width - 54) / 2,
            child: OutlinedButton(
              onPressed: () => vm.topUp(amt.toDouble()),
              style: OutlinedButton.styleFrom(padding: const EdgeInsets.symmetric(vertical: 16), shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)), side: const BorderSide(color: AppColors.border)),
              child: Text('LKR $amt', style: const TextStyle(fontSize: 16, fontWeight: FontWeight.w600, color: AppColors.textPrimary)),
            ),
          )).toList()),
        ]));
      }),
    );
  }
}

// Needed for gradient reference
