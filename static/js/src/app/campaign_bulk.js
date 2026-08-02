var campaignBulk = (function () {
    var selected = {}

    function ids() {
        return Object.keys(selected).map(function (id) { return parseInt(id) })
    }

    function updateButtons() {
		$("#bulkComplete, #bulkResend").prop("disabled", ids().length === 0)
    }

    function checkbox(campaign) {
        var input = $("<input>", {type: "checkbox", class: "campaign-select", value: campaign.id})
        input.attr("aria-label", "Select " + campaign.name)
        return $("<div>").append(input).html()
    }

    function syncVisibleSelection() {
        var visible = $("#campaignTable .campaign-select")
        visible.each(function () { this.checked = !!selected[this.value] })
        var selectedVisible = visible.filter(":checked").length
        var selectAll = $("#selectAllCampaigns").get(0)
        selectAll.checked = visible.length > 0 && selectedVisible === visible.length
        selectAll.indeterminate = selectedVisible > 0 && selectedVisible < visible.length
        updateButtons()
    }

	function complete() {
        Swal.fire({
            title: "Complete selected campaigns?",
            text: "Pending emails for these campaigns will be cancelled.",
            type: "warning",
            showCancelButton: true,
            confirmButtonText: "Complete",
            reverseButtons: true,
            allowOutsideClick: false,
            showLoaderOnConfirm: true,
            preConfirm: function () { return api.campaigns.complete(ids()) }
        }).then(function (result) {
            if (!result.value) return
            var failures = result.value.data.results.filter(function (item) { return item.status !== "completed" })
            if (failures.length > 0) {
                var completed = result.value.data.results.length - failures.length
                Swal.fire("Completed with errors", completed + " completed; " + failures.length + " failed: " + failures.map(function (item) { return item.id + ": " + item.message }).join("; "), "warning").then(function () { location.reload() })
                return
	}

	function resend() {
		Swal.fire({
			title: "Retry failed emails?",
			text: "Only terminal delivery failures will be queued again.",
			type: "question",
			showCancelButton: true,
			confirmButtonText: "Retry",
			reverseButtons: true,
			allowOutsideClick: false,
			showLoaderOnConfirm: true,
			preConfirm: function () {
				return Promise.all(ids().map(function (id) {
					return api.campaignId.resendErrors(id).then(function (response) {
						return {id: id, queued: response.data.queued}
					}, function (response) {
						return {id: id, error: response.responseJSON.message}
					})
				}))
			}
		}).then(function (result) {
			if (!result.value) return
			var failures = result.value.filter(function (item) { return item.error })
			if (failures.length > 0) {
				var queuedWithErrors = result.value.reduce(function (total, item) { return total + (item.queued || 0) }, 0)
				Swal.fire("Retry completed with errors", queuedWithErrors + " queued; " + failures.length + " failed: " + failures.map(function (item) { return item.id + ": " + item.error }).join("; "), "warning").then(function () { location.reload() })
				return
			}
			var queued = result.value.reduce(function (total, item) { return total + item.queued }, 0)
			successFlash(queued + " failed email(s) queued for retry.")
		})
	}
            location.reload()
        })
    }

    function bind() {
        $("#campaignTable").on("change", ".campaign-select", function () {
            if (this.checked) selected[this.value] = true
            else delete selected[this.value]
            syncVisibleSelection()
        })
        $("#selectAllCampaigns").on("change", function () {
            var checked = this.checked
            $("#campaignTable .campaign-select").each(function () {
                this.checked = checked
                if (checked) selected[this.value] = true
                else delete selected[this.value]
            })
            syncVisibleSelection()
        })
        $("#campaignTable").on("draw.dt", syncVisibleSelection)
		$("#bulkComplete").on("click", complete)
		$("#bulkResend").on("click", resend)
    }

    return {bind: bind, checkbox: checkbox}
})()
