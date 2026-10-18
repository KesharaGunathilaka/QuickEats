import 'package:quickeats_customer/models/menu_item.dart';

class CartItem {
  final MenuItem menuItem;
  int quantity;

  CartItem({required this.menuItem, this.quantity = 1});

  int get subtotal => menuItem.price * quantity;
  String get formattedSubtotal => 'LKR ${(subtotal / 100).toStringAsFixed(2)}';
}
