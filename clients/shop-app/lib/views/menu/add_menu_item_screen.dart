import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:provider/provider.dart';
import 'package:quickeats_shop/core/constants/app_colors.dart';
import 'package:quickeats_shop/viewmodels/menu_viewmodel.dart';

class AddMenuItemScreen extends StatefulWidget {
  const AddMenuItemScreen({super.key});
  @override
  State<AddMenuItemScreen> createState() => _AddMenuItemScreenState();
}

class _AddMenuItemScreenState extends State<AddMenuItemScreen> {
  final _formKey = GlobalKey<FormState>();
  final _nameCtrl = TextEditingController();
  final _descCtrl = TextEditingController();
  final _priceCtrl = TextEditingController();
  final _categoryCtrl = TextEditingController();
  final _stockCtrl = TextEditingController(text: '10');
  bool _saving = false;

  @override
  void dispose() {
    _nameCtrl.dispose(); _descCtrl.dispose(); _priceCtrl.dispose(); _categoryCtrl.dispose(); _stockCtrl.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppColors.white,
      appBar: AppBar(
        title: const Text('Add Menu Item', style: TextStyle(fontWeight: FontWeight.w700)),
        backgroundColor: AppColors.white, foregroundColor: AppColors.textPrimary, elevation: 0,
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(20),
        child: Form(
          key: _formKey,
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              _field(_nameCtrl, 'Item Name', 'e.g., Chicken Rice', validator: (v) => v!.isEmpty ? 'Required' : null),
              const SizedBox(height: 16),
              _field(_descCtrl, 'Description', 'Brief description', maxLines: 2),
              const SizedBox(height: 16),
              _field(_priceCtrl, 'Price (LKR)', 'e.g., 350.00', keyboard: TextInputType.number,
                validator: (v) => v!.isEmpty || double.tryParse(v) == null ? 'Enter valid price' : null),
              const SizedBox(height: 16),
              _field(_categoryCtrl, 'Category', 'e.g., Rice, Snacks, Drinks',
                validator: (v) => v!.isEmpty ? 'Required' : null),
              const SizedBox(height: 16),
              _field(_stockCtrl, 'Initial Stock', 'e.g., 10', keyboard: TextInputType.number,
                validator: (v) => v!.isEmpty || int.tryParse(v) == null ? 'Enter valid number' : null),
              const SizedBox(height: 32),
              SizedBox(
                height: 50,
                child: FilledButton(
                  onPressed: _saving ? null : _save,
                  style: FilledButton.styleFrom(backgroundColor: AppColors.primary, shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12))),
                  child: _saving
                    ? const SizedBox(height: 20, width: 20, child: CircularProgressIndicator(strokeWidth: 2, color: Colors.white))
                    : const Text('Add Item', style: TextStyle(fontSize: 16, fontWeight: FontWeight.w600)),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _field(TextEditingController ctrl, String label, String hint, {int maxLines = 1, TextInputType? keyboard, String? Function(String?)? validator}) {
    return TextFormField(
      controller: ctrl, maxLines: maxLines, keyboardType: keyboard, validator: validator,
      decoration: InputDecoration(
        labelText: label, hintText: hint,
        border: OutlineInputBorder(borderRadius: BorderRadius.circular(12), borderSide: BorderSide(color: AppColors.border)),
        enabledBorder: OutlineInputBorder(borderRadius: BorderRadius.circular(12), borderSide: BorderSide(color: AppColors.border)),
        focusedBorder: OutlineInputBorder(borderRadius: BorderRadius.circular(12), borderSide: BorderSide(color: AppColors.primary, width: 2)),
        filled: true, fillColor: AppColors.surface,
      ),
    );
  }

  Future<void> _save() async {
    if (!_formKey.currentState!.validate()) return;
    setState(() => _saving = true);

    final priceCents = (double.parse(_priceCtrl.text) * 100).round();
    final success = await context.read<MenuViewModel>().addItem(
      name: _nameCtrl.text.trim(),
      description: _descCtrl.text.trim(),
      price: priceCents,
      category: _categoryCtrl.text.trim(),
      stock: int.parse(_stockCtrl.text.trim()),
    );

    setState(() => _saving = false);
    if (success && mounted) context.pop();
  }
}
