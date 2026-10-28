import 'package:flutter/material.dart';
import 'package:quickeats_shop/core/constants/app_colors.dart';
import 'package:quickeats_shop/core/di/service_locator.dart';
import 'package:quickeats_shop/models/organisation.dart';
import 'package:quickeats_shop/services/account_service.dart';

/// Lets a merchant add the organisation codes their shop serves. Each code is checked
/// with the server and shown as a chip with the organisation's name.
class OrgCodesField extends StatefulWidget {
  final List<Organisation> value;
  final ValueChanged<List<Organisation>> onChanged;
  final bool enabled;

  const OrgCodesField({super.key, required this.value, required this.onChanged, this.enabled = true});

  @override
  State<OrgCodesField> createState() => _OrgCodesFieldState();
}

class _OrgCodesFieldState extends State<OrgCodesField> {
  final _code = TextEditingController();
  bool _checking = false;
  String? _error;

  Future<void> _add() async {
    final code = _code.text.trim().toLowerCase();
    if (code.isEmpty) return;
    if (widget.value.any((o) => o.slug == code)) {
      setState(() => _error = 'Already added');
      return;
    }
    setState(() { _checking = true; _error = null; });
    try {
      final org = await getIt<AccountService>().lookupOrganisation(code);
      if (!mounted) return;
      if (org == null) {
        setState(() { _checking = false; _error = 'No organisation uses the code "$code"'; });
        return;
      }
      _code.clear();
      setState(() => _checking = false);
      widget.onChanged([...widget.value, org]);
    } catch (e) {
      if (mounted) setState(() { _checking = false; _error = apiErrorMessage(e, 'Couldn\'t check the code'); });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Column(crossAxisAlignment: CrossAxisAlignment.stretch, children: [
      Row(crossAxisAlignment: CrossAxisAlignment.start, children: [
        Expanded(child: TextField(
          controller: _code,
          enabled: widget.enabled && !_checking,
          autocorrect: false,
          textInputAction: TextInputAction.done,
          onSubmitted: (_) => _add(),
          onChanged: (_) { if (_error != null) setState(() => _error = null); },
          decoration: InputDecoration(
            labelText: 'Organisation code',
            hintText: 'e.g. ruh-eng',
            errorText: _error,
            prefixIcon: const Icon(Icons.school_outlined, color: AppColors.textMuted),
            border: OutlineInputBorder(borderRadius: BorderRadius.circular(12), borderSide: const BorderSide(color: AppColors.border)),
            enabledBorder: OutlineInputBorder(borderRadius: BorderRadius.circular(12), borderSide: const BorderSide(color: AppColors.border)),
            filled: true,
            fillColor: AppColors.surface,
          ),
        )),
        const SizedBox(width: 8),
        SizedBox(height: 56, child: FilledButton.tonal(
          onPressed: widget.enabled && !_checking ? _add : null,
          style: FilledButton.styleFrom(shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12))),
          child: _checking
            ? const SizedBox(width: 18, height: 18, child: CircularProgressIndicator(strokeWidth: 2))
            : const Text('Add'),
        )),
      ]),
      const SizedBox(height: 10),
      if (widget.value.isEmpty)
        const Text('Add the code of every university or faculty your shop serves.',
          style: TextStyle(fontSize: 12.5, color: AppColors.textSecondary))
      else
        Wrap(spacing: 8, runSpacing: 8, children: [
          for (final org in widget.value)
            InputChip(
              avatar: const Icon(Icons.school, size: 16, color: AppColors.primary),
              label: Text('${org.name} · ${org.slug}', overflow: TextOverflow.ellipsis),
              onDeleted: widget.enabled ? () => widget.onChanged(widget.value.where((o) => o.slug != org.slug).toList()) : null,
            ),
        ]),
    ]);
  }

  @override
  void dispose() { _code.dispose(); super.dispose(); }
}
