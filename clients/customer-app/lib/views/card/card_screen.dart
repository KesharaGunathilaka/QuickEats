import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:provider/provider.dart';
import 'package:quickeats_customer/core/constants/app_colors.dart';
import 'package:quickeats_customer/viewmodels/card_viewmodel.dart';
import 'package:quickeats_customer/views/card/card_visual.dart';

/// The Card tab: the customer's saved payment card, or a prompt to add one.
class CardScreen extends StatefulWidget {
  const CardScreen({super.key});
  @override State<CardScreen> createState() => _CardScreenState();
}

class _CardScreenState extends State<CardScreen> {
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) { if (mounted) context.read<CardViewModel>().loadCard(); });
  }

  Future<void> _addCard() async {
    final added = await context.push<bool>('/card/add');
    if (added == true && mounted) {
      ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('Card saved. You can now pay for orders with it.')));
    }
  }

  Future<void> _removeCard(CardViewModel vm) async {
    final confirm = await showDialog<bool>(context: context, builder: (dialogContext) => AlertDialog(
      title: const Text('Remove card?'),
      content: const Text('You\'ll need to add a card again before placing an order.'),
      actions: [
        TextButton(onPressed: () => Navigator.pop(dialogContext, false), child: const Text('Keep card')),
        FilledButton(onPressed: () => Navigator.pop(dialogContext, true), style: FilledButton.styleFrom(backgroundColor: AppColors.error), child: const Text('Remove')),
      ],
    ));
    if (confirm != true || !mounted) return;
    final ok = await vm.removeCard();
    if (mounted) {
      ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(ok ? 'Card removed' : (vm.error ?? 'Couldn\'t remove the card'))));
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Payment Card', style: TextStyle(fontWeight: FontWeight.w700)), backgroundColor: AppColors.white, elevation: 0),
      body: Consumer<CardViewModel>(builder: (_, vm, __) {
        if (vm.isLoading && !vm.loaded) return const Center(child: CircularProgressIndicator());
        if (vm.error != null && !vm.loaded) return _ErrorState(message: vm.error!, onRetry: vm.loadCard);

        final card = vm.card;
        return RefreshIndicator(
          onRefresh: vm.loadCard,
          child: ListView(padding: const EdgeInsets.all(16), children: [
            if (card == null) ...[
              const SizedBox(height: 24),
              const _EmptyCardState(),
              const SizedBox(height: 24),
              SizedBox(height: 52, child: FilledButton.icon(
                onPressed: _addCard,
                icon: const Icon(Icons.add_card),
                label: const Text('Add a card', style: TextStyle(fontSize: 16, fontWeight: FontWeight.w600)),
                style: FilledButton.styleFrom(backgroundColor: AppColors.primary, shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12))),
              )),
            ] else ...[
              CardVisual(card: card),
              const SizedBox(height: 16),
              Container(
                padding: const EdgeInsets.all(14),
                decoration: BoxDecoration(color: AppColors.secondary.withOpacity(0.08), borderRadius: BorderRadius.circular(12)),
                child: const Row(children: [
                  Icon(Icons.verified_outlined, color: AppColors.secondary),
                  SizedBox(width: 10),
                  Expanded(child: Text('This card is used for all your orders.', style: TextStyle(color: AppColors.textPrimary))),
                ]),
              ),
              const SizedBox(height: 24),
              SizedBox(height: 50, child: OutlinedButton.icon(
                onPressed: _addCard,
                icon: const Icon(Icons.swap_horiz),
                label: const Text('Use a different card', style: TextStyle(fontWeight: FontWeight.w600)),
                style: OutlinedButton.styleFrom(shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12))),
              )),
              const SizedBox(height: 8),
              TextButton.icon(
                onPressed: () => _removeCard(vm),
                icon: const Icon(Icons.delete_outline),
                label: const Text('Remove card'),
                style: TextButton.styleFrom(foregroundColor: AppColors.error),
              ),
            ],
          ]),
        );
      }),
    );
  }
}

class _EmptyCardState extends StatelessWidget {
  const _EmptyCardState();
  @override
  Widget build(BuildContext context) {
    return Column(children: [
      Container(
        width: 96, height: 96,
        decoration: const BoxDecoration(color: AppColors.primaryLight, shape: BoxShape.circle),
        child: const Icon(Icons.credit_card_off_outlined, size: 44, color: AppColors.primary),
      ),
      const SizedBox(height: 20),
      const Text('No card added yet', style: TextStyle(fontSize: 20, fontWeight: FontWeight.w700)),
      const SizedBox(height: 8),
      const Text('Add a debit or credit card to pay for your orders.',
          textAlign: TextAlign.center, style: TextStyle(fontSize: 14, color: AppColors.textSecondary)),
    ]);
  }
}

class _ErrorState extends StatelessWidget {
  final String message;
  final Future<void> Function() onRetry;
  const _ErrorState({required this.message, required this.onRetry});
  @override
  Widget build(BuildContext context) {
    return Center(child: Padding(padding: const EdgeInsets.all(24), child: Column(mainAxisSize: MainAxisSize.min, children: [
      const Icon(Icons.cloud_off, size: 40, color: AppColors.textMuted),
      const SizedBox(height: 12),
      Text(message, textAlign: TextAlign.center, style: const TextStyle(color: AppColors.textSecondary)),
      const SizedBox(height: 12),
      OutlinedButton(onPressed: onRetry, child: const Text('Try again')),
    ])));
  }
}
