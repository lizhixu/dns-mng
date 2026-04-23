# hurricane.md

## 登录

- **API 地址**：`POST https://dns.he.net/`
- **认证方式**：`email=your_email@example.com&pass=your_password&submit=Login%21`
- **支持格式**：`application/x-www-form-urlencoded`
- **说明**：
  - 需要从响应中判断是否登录成功
  - 需要保存 `cookie` 供后续请求使用
  - 需要解析 HTML 中域名相关节点里的 `name` 和 `onclick` 中的 javascript 参数

- **部分关键响应**:

```html
<table width="100%" id="domains_table" class="generic_table" border="1" cellpadding="0" cellspacing="0">
    <thead>
        <tr>
            <th><img src="/include/images/link_go.png" alt="Open Link"/></th>
            <th><img src="/include/images/pencil.png" alt="Edit" /></th>
            <th>Active domains for this account<img src="/include/images/help.png" onclick="$('#dialog_active_domains').dialog('open')" style="cursor: help;" alt="help" /></th>
            <th><img src="/include/images/delete.png" alt="Delete" /></th>
        </tr>
    </thead>
    <tbody>
        <tr>
            <td style="cursor: pointer;">
                <img class="Tips"
                     title="Open URL::Open example.domain.test in a new window."
                     alt="go" src="/include/images/link_go.png"
                     onclick="window.open('http://example.domain.test','example.domain.test')" />
            </td>
            <td style="cursor: pointer;">
                <img class="Tips"
                     title="Edit Zone::Use this option to edit the zonefile. You would use this if you wanted to add or remove subdomains, etc"
                     alt="edit" src="/include/images/pencil.png" name="example.domain.test"
                     onclick="javascript:document.location.href='?hosted_dns_zoneid=1000000&menu=edit_zone&hosted_dns_editzone'" />
            </td>
            <td width="100%" style="padding-left: 3px;">
                <span>example.domain.test</span>
            </td>
            <td style="cursor: pointer;">
                <img class="Tips"
                     title="Using this option will PERMANENTLY remove the zone from your account."
                     alt="delete" onclick="delete_dom(this);" name="example.domain.test" value="1000000" src="/include/images/delete.png" />
            </td>
        </tr>
    </tbody>
</table>
```

- **解析要求**：
  - 提取域名：`name="example.domain.test"`
  - 提取编辑链接中的参数：
    - `hosted_dns_zoneid=1000000`
    - `menu=edit_zone`
    - `hosted_dns_editzone`

---

## DNS 解析记录

- **API 地址**：`GET https://dns.he.net/?hosted_dns_zoneid=1000000&menu=edit_zone&hosted_dns_editzone`
- **支持格式**：`application/x-www-form-urlencoded`
- **说明**：
  - 需要使用登录后保存的 `cookie`
  - 需要解析每条 DNS 记录
  - 注意解析删除按钮 `onclick` 中的参数
  - 报错信息也在返回 HTML 中，需要解析 DOM 并进行 HTML 实体反转义

- **部分关键响应**:

```html
<table class="generictable" width="100%" border="1" cellpadding="0" cellspacing="0">
    <tr>
        <th class="hidden">Zone Id</th>
        <th class="hidden">Record Id</th>
        <th style="width: 25px;">Name</th>
        <th style="width: 25px;">Type</th>
        <th style="width: 25px;">TTL</th>
        <th style="width: 25px;">Priority</th>
        <th style="width: 25px;">Data</th>
        <th style="width: 25px;">DDNS</th>
        <th style="width: 25px;">Delete</th>
    </tr>

    <tr class="dns_tr_locked" id="9000000001" title="These items are not editable." onclick="lockedElement('token_example_1')">
        <td class="hidden">1000000</td>
        <td class="hidden">9000000001</td>
        <td width="95%" class="dns_view_locked">example.domain.test</td>
        <td align="center"><span class="rrlabel SOA" data="SOA" alt="SOA">SOA</span></td>
        <td align="left">172800</td>
        <td align="center">-</td>
        <td align="left" data="ns1.he.net. hostmaster.he.net. 2025010101 86400 7200 3600000 172800">ns1.he.net. hostmaster.he.net. 2025010101 86400 7200 3600000 172800</td>
        <td class="hidden">0</td>
        <td></td>
        <td align="center" class="dns_delete_locked">
            <img src="/include/images/lock.png" alt="locked" title="This element is locked and may not be deleted."/>
        </td>
    </tr>

    <tr class="dns_tr" id="9000000002" title="Click to edit this item." onclick="editRow(this)">
        <td class="hidden">1000000</td>
        <td class="hidden">9000000002</td>
        <td width="95%" class="dns_view">example.domain.test</td>
        <td align="center"><span class="rrlabel NS" data="NS" alt="NS">NS</span></td>
        <td align="left">172800</td>
        <td align="center">-</td>
        <td align="left" data="ns1.he.net">ns1.he.net</td>
        <td class="hidden">0</td>
        <td></td>
        <td align="center" class="dns_delete" onclick="event.cancelBubble=true;deleteRecord('9000000002','example.domain.test','NS')" title="Click to delete this record.">
            <img src="/include/images/delete.png" alt="delete"/>
        </td>
    </tr>

    <tr class="dns_tr" id="9000000010" title="Click to edit this item." onclick="editRow(this)">
        <td class="hidden">1000000</td>
        <td class="hidden">9000000010</td>
        <td width="95%" class="dns_view">example.domain.test</td>
        <td align="center"><span class="rrlabel A" data="A" alt="A">A</span></td>
        <td align="left">300</td>
        <td align="center">-</td>
        <td align="left" data="203.0.113.10">203.0.113.10</td>
        <td class="hidden">0</td>
        <td></td>
        <td align="center" class="dns_delete" onclick="event.cancelBubble=true;deleteRecord('9000000010','example.domain.test','A')" title="Click to delete this record.">
            <img src="/include/images/delete.png" alt="delete"/>
        </td>
    </tr>
</table>
```

- **记录解析要求**：
  - 每条记录至少提取：
    - `zoneId`
    - `recordId`
    - `name`
    - `type`
    - `ttl`
    - `priority`
    - `content`
    - 是否锁定
  - 删除按钮中的 `onclick` 需要解析：
    - `deleteRecord('9000000010','example.domain.test','A')`
    - 提取：
      - `recordId=9000000010`
      - `name=example.domain.test`
      - `type=A`

---

## 添加解析

- **API 地址**：`POST https://dns.he.net/index.cgi`
- **支持格式**：`application/x-www-form-urlencoded`
- **请求示例**：

```txt
account=&menu=edit_zone&Type=A&hosted_dns_zoneid=1000000&hosted_dns_recordid=&hosted_dns_editzone=1&Priority=&Name=www&Content=203.0.113.10&TTL=300&hosted_dns_editrecord=Submit
```

- **字段说明**：
  - `Type`：记录类型，如 `A` / `AAAA` / `CNAME` / `TXT`
  - `hosted_dns_zoneid`：Zone ID
  - `Name`：主机记录
  - `Content`：记录值
  - `TTL`：TTL
  - `Priority`：MX 等记录类型时需要

---

## 编辑解析

- **API 地址**：`POST https://dns.he.net/?hosted_dns_zoneid=1000000&menu=edit_zone&hosted_dns_editzone`
- **支持格式**：`application/x-www-form-urlencoded`
- **请求示例**：

```txt
account=&menu=edit_zone&Type=A&hosted_dns_zoneid=1000000&hosted_dns_recordid=9000000010&hosted_dns_editzone=1&Priority=-&Name=app.example.domain.test&Content=203.0.113.11&TTL=300&hosted_dns_editrecord=Update
```

- **说明**：
  - 编辑时必须带上已有的 `hosted_dns_recordid`
  - `Type` 建议保持与原记录一致
  - `Name` 可能是完整域名，也可能是相对主机名，需根据实际页面行为兼容处理

---

## 删除解析

- **API 地址**：`POST https://dns.he.net/index.cgi`
- **支持格式**：`application/x-www-form-urlencoded`
- **请求示例**：

```txt
hosted_dns_zoneid=1000000&hosted_dns_recordid=9000000010&menu=edit_zone&hosted_dns_delconfirm=delete&hosted_dns_editzone=1&hosted_dns_delrecord=1
```

---

## 报错信息

- 需要从返回 HTML 中解析错误节点
- 注意对 HTML 实体进行反转义

**示例**：

```html
<div id="dns_err" onclick="hideThis(this);">CNAME at zone apex is not allowed. (rfc1912 &amp; rfc2181)</div>
```

**解析后错误信息应为**：

```txt
CNAME at zone apex is not allowed. (rfc1912 & rfc2181)
```

---

