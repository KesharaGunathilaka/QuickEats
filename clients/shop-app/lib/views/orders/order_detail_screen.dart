import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:quickeats_shop/core/constants/app_colors.dart';
import 'package:quickeats_shop/core/constants/order_status.dart';
import 'package:quickeats_shop/models/order.dart';
import 'package:quickeats_shop/viewmodels/order_queue_viewmodel.dart';

class OrderDetailScreen extends StatefulWidget {
  final String orderId;
  const OrderDetailScreen({super.key, required this.orderId});
  @override
  State<OrderDetailScreen> createState() => _OrderDetailScreenState();
}

class _OrderDetailScreenState extends State<OrderDetailScreen> {
  final _rejectController = TextEditingController();

  @override
  void dispose() {
    _rejectController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Consumer<OrderQueueViewModel>(
      builder: (context, vm, _) {
        final order = vm.allOrders.where((o) => o.id == widget.orderId).firstOrNull;
        if (order == null) {
          return Scaffold(appBar: AppBar(), body: const Center(child: Text('Order not found')));
        }

        final statusColor = AppColors.forStatus(order.status.value);
        return Scaffold(
          backgroundColor: AppColors.surface,
          appBar: AppBar(
            title: Text('#${order.shortId}'),
            backgroundColor: AppColors.white,
            foregroundColor: AppColors.textPrimary,
            elevation: 0,
          ),
          body: SingleChildScrollView(
            padding: const EdgeInsets.all(16),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                // Status banner
                Container(
                  padding: const EdgeInsets.all(16),
                  decoration: BoxDecoration(color: statusColor.withOpacity(0.1), borderRadius: BorderRadius.circular(12)),
                  child: Row(
                    children: [
                      Icon(Icons.circle, size: 12, color: statusColor),
                      const SizedBox(width: 8),
                      Text(order.status.label, style: TextStyle(fontSize: 16, fontWeight: FontWeight.w600, color: statusColor)),
                      const Spacer(),
                      Text(order.waitTimeFormatted, style: TextStyle(fontSize: 13, color: AppColors.textSecondary)),
                    ],
                  ),
                ),
                const SizedBox(height: 16),

                // Pickup code
                if (order.pickupCode.isNotEmpty)
                  Container(
                    padding: const EdgeInsets.all(20),
                    decoration: BoxDecoration(color: AppColors.white, borderRadius: BorderRadius.circular(12), border: Border.all(color: AppColors.border)),
                    child: Column(
                      children: [
                        Text('PICKUP CODE', style: TextStyle(fontSize: 12, fontWeight: FontWeight.w600, color: AppColors.textMuted, letterSpacing: 1)),
                        const SizedBox(height: 8),
                        Text(order.pickupCode, style: TextStyle(fontSize: 36, fontWeight: FontWeight.w800, color: AppColors.primary, letterSpacing: 4)),
                      ],
                    ),
                  ),
                const SizedBox(height: 16),

                // Items
                Container(
                  padding: const EdgeInsets.all(16),
                  decoration: BoxDecoration(color: AppColors.white, borderRadius: BorderRadius.circular(12), border: Border.all(color: AppColors.border)),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text('Items', style: TextStyle(fontSize: 14, fontWeight: FontWeight.w600, color: AppColors.textPrimary)),
                      const SizedBox(height: 12),
                      ...order.items.map((item) => Padding(
                        padding: const EdgeInsets.only(bottom: 8),
                        child: Row(
                          children: [
                            Container(
                              width: 28, height: 28,
                              decoration: BoxDecoration(color: AppColors.surfaceAlt, borderRadius: BorderRadius.circular(6)),
                              child: Center(child: Text('${item.quantity}', style: TextStyle(fontWeight: FontWeight.w600, fontSize: 13))),
                            ),
                            const SizedBox(width: 12),
                            Expanded(child: Text(item.itemName, style: TextStyle(fontSize: 14, color: AppColors.textPrimary))),
                            Text('LKR ${(item.subtotal / 100).toStringAsFixed(2)}', style: TextStyle(fontSize: 13, color: AppColors.textSecondary)),
                          ],
                        ),
                      )),
                      const Divider(),
                      Row(
                        children: [
                          Text('Total', style: TextStyle(fontSize: 15, fontWeight: FontWeight.w700)),
                          const Spacer(),
                          Text(order.formattedTotal, style: TextStyle(fontSize: 16, fontWeight: FontWeight.w700, color: AppColors.primary)),
                        ],
                      ),
                    ],
                  ),
                ),
                const SizedBox(height: 24),

                // Action buttons
                if (order.status.nextShopAction != null)
                  SizedBox(
                    height: 50,
                    child: FilledButton(
                      onPressed: () async {
                        final success = await vm.updateStatus(order.id, order.status.nextShopAction!);
                        if (success && mounted) setState(() {});
                      },
                      style: FilledButton.styleFrom(
                        backgroundColor: statusColor,
                        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                      ),
                      child: Text(order.status.nextActionLabel, style: TextStyle(fontSize: 16, fontWeight: FontWeight.w600)),
                    ),
                  ),

                if (order.status == OrderStatus.confirmed || order.status == OrderStatus.accepted)
                  Padding(
                    padding: const EdgeInsets.only(top: 12),
                    child: SizedBox(
                      height: 50,
                      child: OutlinedButton(
                        onPressed: () => _showRejectDialog(context, vm, order),
                        style: OutlinedButton.styleFrom(
                          foregroundColor: AppColors.error,
                          side: BorderSide(color: AppColors.error),
                          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                        ),
                        child: const Text('Reject Order', style: TextStyle(fontSize: 16, fontWeight: FontWeight.w600)),
                      ),
                    ),
                  ),
              ],
            ),
          ),
        );
      },
    );
  }

  void _showRejectDialog(BuildContext context, OrderQueueViewModel vm, Order order) {
    showDialog(
      context: context,
      builder: (_) => AlertDialog(
        title: const Text('Reject Order'),
        content: TextField(
          controller: _rejectController,
          decoration: const InputDecoration(hintText: 'Reason for rejection', border: OutlineInputBorder()),
          maxLines: 3,
        ),
        actions: [
          TextButton(onPressed: () => Navigator.pop(context), child: const Text('Cancel')),
          FilledButton(
            onPressed: () async {
              if (_rejectController.text.trim().isEmpty) return;
              await vm.rejectOrder(order.id, _rejectController.text.trim());
              if (context.mounted) Navigator.pop(context);
              _rejectController.clear();
            },
            style: FilledButton.styleFrom(backgroundColor: AppColors.error),
            child: const Text('Reject'),
          ),
        ],
      ),
    );
  }
}
