package output

var HTMLTemplate = `
{{define "style"}}
body {
  margin: 0;
  padding: 0.5rem 1rem;
  font-family: sans-serif;
  color: #111827;
}

a {
  color: #3b82f6;
}

.metadata {
  margin-bottom: 1.5rem;
}

.metadata ul {
  list-style-type: none;
  padding: 0;
}

.metadata ul li {
  margin-bottom: 0.5rem;
}

.metadata .label {
  display: inline-block;
  font-weight: bold;
  margin-right: 0.5rem;
  width: 8rem;
}

.warnings {
  margin-top: 1.5rem;
}

table {
  border-collapse: collapse;
  border: 1px solid #6b7280;
}

th, td {
  padding: 0.25rem 0.5rem;
  text-align: left;
}

td.name {
  max-width: 32rem;
}

td.monthly-quantity, td.price, td.hourly-cost, td.monthly-cost {
  text-align: right;
}

tr.group {
  background-color: #e0e7ff;
}

tr.resource {
  background-color: #e5e7eb;
}

tr.resource.top-level {
  background-color: #6b7280;
  color: #ffffff;
}

tr.tags {
  background-color: #6b7280;
  color: #ffffff;
  font-size: 0.75rem;
}

tr.tags td {
  padding-top: 0;
}

tr.total {
  background-color: #ffdfb9;
  font-weight: bold;
}

tr.total td {
  padding-top: 0.75rem;
  padding-bottom: 0.75rem;
}

.arrow {
  color: #96a0b5;
}
{{end}}

{{define "faviconBase64"}}
iVBORw0KGgoAAAANSUhEUgAAAMAAAADACAMAAABlApw1AAAABGdBTUEAALGPC/xhBQAAAAFzUkdCAK7OHOkAAAAJcEhZcwAAhOAAAITgATg6g3cAAAGDUExURUdwTHZZw8dzrm5YxK1gun1awnFZxKpfuq9gubJgua93rrF0q9aTm4hbv6Jeu21ZxG1ZxMl5qaZfu+Ssj+OqkMt8qW5ZxOKpkKRfu8l4qqFfu6Beu+OpkeOrj8yAp5D/yf+k/7NduP///7NhuJhdvbZhuKxgubhht49dvr5itqVfurpht5Vdvb9ktcFpsqJfu6Beu8JrsZ1evIZbwJ5evJpevMBmtLFguNiWm5Ndvqpguq9gucNtsH1bwcFos41cv4FbwYtcv8p6qst8qLtit8Rvr7xit9ycmMx/p4lcv3hawsh2rM+DpdiUnN2el9uamdGIo9mXmt+ilc6BptaRntCGo9OMoKhfuoRbwHpawt6glqNfu8VxrteSnadfuuCklJJdvsd0rXVZw+GmktKKobRguNWQn9SOn3JZw8l4q9uamqRfu9SNoOSskIJbwW5ZxPfv9uOqkdKJos+EpOKokvLg69yx2uS71syv3tqdqOnFy/DW3OCtt82Du9OWyNqjyrua1oHj538AAAAfdFJOUwD+/v7+/v/+/v4gEFxchofphO9/2dnDlbm74M+/sJ+SqbCHAAAe20lEQVR42rWda1dUx9KAN8A4M6oxycn9nPd8mAGQmyCIyB2SYRRUISqAoojqgAAqghJMNDk//e2u6ktVdQ84G9jmS1gLVj+rrt1VXZ0k8jt34auffvm2ZVh9AwMD+WKxUqk0Nze3qq/UUSqVy21tbV3qu3ZtbGzs8uUr6htR39TU1GP1PYPv1q1bGxsbv6rvN/XdVN/9+/dv3769p775+bm53d2X6ltY2F5fX19T3xP13blz57r6fv99c3NWfTfUd/fu3UePHt1T31P1ffefH8+fTY74zn3102JLtiU3rP5T/xRBMV90AK0lBqAQLjuCqSlEsAARAkSY1wRzLzXCwvbCukZ48OABANzR61cEZv2S4MX7F++/O5Th3FffrixmF4eGWvQ3rBA0QREI9Po7rATaLMDY5SuAgABq/fUUYAPX7wDu6/WDCJQQXi4oEWgZrK09WDMEKAGQAQeA9avv/fv3736sivDVt+3tK2r5i1m1/MYW0KE8ADQ3O4JyiQGMgQQMgBOBB3ASuGkANMEcKtHCSwOwRgG0EGY3Z2/M2vVTAg3w7t2/oss/++92tf6VxcXskBPB8LCygQEvglYQAVEiq0NUBPX11giq69D8nLWCBSAAJXrilEjrUGgFL1AGiuC7iBAufLvVvgUEDiAHOpQ3ACgCsGMCMGYQjBWr//T6QQS3EOBXpkMIYO1YAWzHAByBkQESvHgKIlAEr9+fD9RneVktX8tgMZtVCI1KiZQAchpggJixViIrAmMFBsCZsZIBNWNrBfdvWoI9FMEuc0SGAB2RBpg1AEIEav3v371+91qo0Vfd3d3LsH6tRMoKhkAC4IcG8sSTdpRQAhZAK9EV4okeEyu4FVqBWT/KAK3A6pBe/xMLsOkJHnk71utHHXr3mhNc6OnpXt4yBIuKwJlxTgkgn89TEZSdI3JKZGLBCPFE3I5v/kbMeM/o0C7q0DYhMI5IeSIugkdMBO+0CF4TLTr7bU+3IljWRqAJhoay2UbvSYvMEZU6WCwgOuTsuD4eC6wZaz8EsQB0yHqiB0+kFVgzvktCgUV4rT5vyb/0KAn0LKtPA7SvuFCArnRAWgFTIhLMPEH9s1vPogDWjufQChZ0MFtgIkCCTRfNqBlTGbx+/Z0zgKtXNYGygi1jBUNA0AhmMGzCsY8FHQjQ5mOBRhi54uxYSSDqSvXyiRlYT8qVSOiQjMbWEWkCYwZnezVAd7claNdGAFYAIgABFNEKNEFHaweLBc6T+ljgzUDEgpssI7KxYGF9YR08qSX43TiiGzekI3qBIjBW8O4cCqDXSQB0SAkhO5TVBEoEmBGhFVScIxL5hMzp6p0ZbESU6PbebZtQGIDtBRcLwowoNOMXL95REWgBWIJuA7Cig5lCaGwxGR2EY5pQxGIB9aT19VVTuvu3vSOyIoCUziuRM+NZZsb3aD6hzRhEcKFXfUaJTDRTZpA18dgYQT5qxm3WD43JpLTexGMrAe6I9lg4NhnR+toDFsw2Z6t60hfMCv5tAXoowOIQmkGjFkHe5KSVKuE4BHAJBdUh6koNwi460gUWjSN2bHK6p0+FDP6jNKi/3wN0+4RCeyK1/kZtBcOwsfE6hATEEV2+TIPZY1SiQ7YFNp2wSSm40nUE8Gk1ZNWzFuCeT0p9OH59LrnQ7wi6QQYAsKiswAQD2Jrli5rAAbR2CDu+bK1gampk6nEkGkeS0j2TT0BOur3tRPCkSji+FySlr3U4/goBjBUsWzOGrFq7oZZcbtjlpNYPtZK9JQEYucLziXoiAk7gEgrniCLBjNkxSUqdFSiAH5OfmhCg11qBSSh0LBgyRqC2BTqYVUI77hLBzEdjvzdTAMHWbM/Hgl0JQERgdsd3YxkRiuD/kl/U+r0IdEa0tYUE2SGMBcIRRXfH2hHZjc2IBaivtr3X8diJwDiideKInrh8YjMAEATfJU39joDGAuWIhkAEjY0tmJXadKJ6RiRE8JgfUPwWOaAAO365a3bHa2s+p+NZ9V2vRMwPqXgMAMaMeSxYtEpkzBjsmIuA7gvI3nKEA0hPKq3Ai2CdmTEjuEv39y8IgQHoN57UE/xvp+A+LYHcQIF8mFD4/9/5R6TVj6undADw6cNOYefDX39++mgcESoRJYiE40csqzYATVYCLhyDDZD1F/T6BUCHPuUiP9i5FgAE0cxvbG7/QX71w5+f4IxoWxuBs+PrbntPk9LggEIDECUysUAD0OWCFuW5BHQooD+JHdNpAuFJUQJ0/YD/1ycFsOCVyIVjsbGxOvTUhoJ3GqCJGIHWIvCkAcCwAGhlEiiM+WDmDZmcclEdkutHho/khMXnE5ubMRG8cEpEAHqvuowoIoGcBFAbG/oTTEphfw9mzDIiHgti69ffX5/MMR07n9iMicDvzJJOQ+AyIrM1o38ZY4GUQGsAMCZ3x3Rj4wCqrV+bw8c1HwqoK70hj+nczgwl0GQdkU9KOYAWQQBQigAohBFPUP+MH3JpgEPWr6XwUewtWUYUHHJpCXQyM3axgKuQ3hjQn2As4AD0mM5vbHwsQEd0+PrV9yfZF/zujuliR72QlCqATq9DvahD+oCCSQD2locDdAU7M59QkKz0yPUrc/74gO/vQ0/qXSkF4PnEERJojkgAD6svMwKS0mmAL1i/FoKPBdeNH4ofUCiCAMDubCRAy9EAY2NkZzNFSzZq/UDwZevXlkAzosghl7djDUAc0dVqAMoRHQ7gzhmv2J2N2N/XsH6tRpGULpIRvXAA/TYpNa6UAeApFwfQCBygC1O6IBw7R/Tl61cEn5wOcRGwfOK9ByDRDPMJBgA6JABiEojvC4wS1bJ+9X0iOhQ/6n1KJcDNWALoE6IjJXAtYgZEh2pcvzLlSOEyrNgkg52dyEDtuLunhwMMCQlUqqgQ6NDliCOqff1aBj6aRR0ReKHBQSECY8f0L8EhVyMDAAIBIEs2tOCRYv3KDnzh0m9sbCgwMkCATptPIMBVAYDnjPQnxSiALfsJAhXL0qxf+6I78X3BvUcEoHOQmLHfF9A/hHtLDlARALGyn8uq060fCe74I6JY1UxJYJB4UhcMBMBiDKCZAdDKq0hK065fRbRqe0t3RpSsDsatgAHoQ7ohBgAEHKCNA3g/lH79yhXR03YSzJwVAMBgEI6vBgCLWfoTLPtFAYKq2XHWr5XIiGAzvrlMVldx/TKlo38FimZCAsUQgBJ4ERxr/WqPc3hSqgE8QT8xYwYAZ9UCQClRKIGgg2LkmOtX0SCyNSPhOLlkAZrY3vIqA9CnXFyFigNKh+hPRAMF6JAiOPb6lRIdesKiJSBEEJEA1JzoTwag5iQBwtrx8dev7diF482wdqwloAg6hRn3MglA+Z4D5HUDAgMoh9X7K/uFk/iuu6R0U4aCe1EAkMHhAHmo2DCAEuvkgn3Byazfi2A2khElly45JXIA/TGARSEBARD2H1z+o3BC3+adowAGpQ719jIAIGAAUC9gAPSwGnRo/6TW70QQK1wqAE5gdYgDaAIBoL5QAsSVntz6VSyo6ohAApdiViAAFAH9CTZQcAnwqtkJrl/HAkHg7VhLwIvgMAAmgZyqmgmADiaDE10/6pDYmRkRWIBAiejvbwUSyOWkCnWQNqKuE15/YYeV/W7QnQ0AeFfaVBWAn1djNx0veRARnPD6C4WPsdN2CGbJuNMhLgL661h55QDDQgK6Fw0ITmP9mFXTtNpm1Y8QwBO4ExYGsCUlkAsBVM3JNFCc/PoLf/ma0+9chwyAS+kgI2oSEgiLTthSKgBUV68GOIX1F3bcAcUsr9gogPG4GTMJRAC0EoVVMyWC01i/DsZV2gEVgCZYZSLQCFwCkaJTUHjF4vfprF+fEV33LSCkqzeZHB93OkQzIgYAOiSKTiGAtoJTWr87piP5BIogmZx0VsCiGQfQ/QdMArpyHBSdOjpOa/3Kiv1RLzspVRKYFI6oMwIgq2YtsapZR+uprR8A7sSSUiWBAAAI6G/3xAByw0HV7PTWr9zQk+CkVLeAgAoZO1YAJKfjAFqHZM0mqJqd4vo9gEyrtQQmiSuNAmAvGgdoDABOc/2FQtBYbQGmQQTeEVmCHSkBBgCdUFyFTnf9hfg1IQQICZQN0HabSNWsUXVCMQmc8voLvPrtRZBMTzsdAiXqlJvLsCNzyHTTQWMydIYXK/u+q7dU4tcLPv/99x8fTkICXolISykCeEe0OihP220DxbK9IQGtXO6WEHZW7xciLaW2rXdMHbB83t85EQnckRUbBcCsYJUn1bTy2u1aSlVneDbrbgkN6PUXSEtpie+OsYdlZOSfneMB8LZee8alvJC3ApERyR6WLduYrPsZobEaLkjkB/ZN1YzdlmMnLFeOfU73QHTTWRFoCUyraDxpAaocVtueUtChFdMVa5rb9wsEoCPoyPQ9pVOf0wtB9iVbAuWFpienSSwYXO0cFErUazq5uv1NrSFyXXHf1S3J/YLILRt9WP35wzEAWCsXAsyiBJgjIodczo5ZP+OKb8lszNn1m7JftLHaWYEieJaWQDa3280xGDGJBSylaxJtRLY1HG9qoQz2aeGVO6KuLtncPpWeIOjOn8U+IrQByIfCWNDfFDZCbdnrBdkh3Ua0T6pm5JpQ2V6QCG9qpbQD0lPKghmqEDLwbUGnrLwSJbJXRrP7pGYj+pLLVWpOjz+nlcBaLBonfdPcCnhKF0azdn3EsmiMYF/WLaPBjLVyjaSsmmFLaXBTSwH09RmA+MZGSqDd3S/ILu7Lqhm9aha5JmSvmu2kAiBXPEw41nZsAaapGZt8ojMSjrEveaV9Mahb5glAqVQSCQUrXP6dzgbWpQg0gQYIRSDPGa0ZuzsqUPXjB+5cBKUyvWTjGqEMwYd0KrROLlzacEwlQLf3nbEGCm7GsuxXZGbAMyLqiEZSiSC46GRyOisBMOJJecoV7+rdskbAqmYqJ6r4YMDvvI7J1vDaRbC9LlrDzdYsWerzMogBiDYiFsx4zUZc8WBX12UHRe2OyFz9fsBv7W46AJ2Ujvu0WjRQ9EasYIWXvvGqWXMzTUrb2sq88nrZ3hh9lhZAXpZDFbKxYDzIiDpFOyA05y+3H8QAsHJJonGJh+Mxeue1Zh1yd1SePGA6pCSwRHRoslo4pk2xWwcFvLXLjnvxymiFxwJ+z+k4wcwN0RCxIFlSXwAQVL97qRkfVC37FUlCYfKJrkg4VjndP7UCvNTXnMSd1+sA0GdFYByR39iEwQAIDkzZr50D5MTNb7Bjvr0n0exzOgmsyxEUVgIOQeZ0TYEOHRxS9rOhILg4HSSl9bVLgE0BcWdEyRu//hhAkNIduMJrrGrGgllrRym4+e1EUDPAS7x8vy48KUhgiRCMxwisCHrN+gvLoQp5Ars7pnZMsmqIxiMpAJQj2g5u+yVv3niAycmjHNGBK/sFV7Xc5fvmSnN0ms8Y06E0AKBD/LofAvQ5RxRJSmk8PvBFJ6lCZpqPuDhdDmbhWB1KAfBSjqBwAOhJ+6bp+QRPSpHggJX9tgIAM8KBDXAoRXY26VQIb/thSuejmVGhpaWYHYtDrgNWeN2SKuRu39OtWYc95LrmjulQBLUC4EirbZ5QaIAJFIFP6caZCIgrPWBFp9h1xRyMwskXeUoXm2NyJR0AuzJ6BwgsAE1Kx8W+oErZL1Y1QzMuVmIA/J5TGgAzVmyN7o6TiQmjRJiUUhms8rNeUXRSDRQSIBfGgvC03TiiWgHmLMF2ACCUiBbNSMFDlP30VTMBoCdCKYJiMZ8XIygwKb1GdzYpAcwMinWXEd3xAEvh5pI6ov6g7LcV3PazVsA8Ke6Ou9rYUe/IldoB5tw4ojWyNXMAPh8SOjQYLbxGyn5wY3Q4j1bQzAaxsGY6DAY1A9gJDtsLa3RrloxqAKFDoSsNCq8xgEZdvc/l8/k8meDQ0SHPqkGHalaheTvTyvkhDAUagCvRJDdjZQUHSCBVSNw1c0UnGEdUrO5Jx9JIYMeOD5AToZLR0cAKfEYEOnRQQBlwCfQIAKg5DRuAPLeC2FCuWgHm5sgcE0KAABMcYHqSHFar+FUFoFtWjhsbbT5RheCaT0pTS4CHgieoQoKAnrDo+IuxgP69q+F1RTMXLZcbNkM+/QiKjkjNqWaAeTLecMEfcikJcBGIjAjyh8EogLxrlvVGMCCH+fCynxZB7RLAUTILaMhrVgYawFlBX5DSYf7zJQD6qlkj1JxwHFG+EtQtvRKlAXA6RJJSCeAlMI1mbPI3dKUBQI8AGMJhPuqf2RZUnA6FVb9aAfboeEOSk2qAjCXo4yndpM0/MaFgAPHrim5M6UCej6bzOmT3ZmlsYH7ORmM/QyOB9Y8KT4qxwOXPoQTwgIIBsFEyOI6ouRIJx2ZrlgKAToRyZpyo9Y9mJlw4dlZA1l/AeMwAeiUAFM3MgEaaVrfKWGBueKRRIU3w0gxUsgMak4eKIJNxjshmRNNk/YXBECB6288MaDSelOUTNqdzjigVwBwVASpR8jAzKkUARvA/8tuYlEqA8LZf1trxsExKdQMCz+nSGfG8G+Zjo7ECyKAI3vicTgN8YAARFQouy8GQT2iEyuVykWhcYhMmawawo93szmzbSgBEQFI6o0Q7QgKDAkDeNfONUNhBkXciaI2eM6YD8OOGt83WTAFoI9AEbyboIRcFwKxUAPSGt/30qNus64NyM7ebI6ft19IAAIHNqk1arQAe+lDgAfokwCoD6A+uK0LFw+QTME9JA5C9pRzKVSsAHdC4S6IZAmS8Etl0ggFckhIIr2qtmBmZNpgN42U/tzXrIHPR0gDQab10b6kBtAgmnBVgOO47XIUit/3skE8y6raoZ93KpNToUM0SuO2H6710+wKUgDbjDIpgyQUzKYHVVQ4Que23Ak04jXbmNgw4ZHtLqkS1S8DNPd81I6tBiZKHMyABjAU+q176AgB5228RXSm0A+pgloe9ZbTgkQogmA64oFVoBkWQyRgdsq6UA2gEARDeNYM+qEXcFtiEYkBU7x1BGgn4ueeYVmtHhAAPM1aHbDiWEvgiADcv2WVE2owJAJRsyukA3KRYPnk+mUERZIwrnbAbmzQSIJPn7YRJNqbU1o61HdcOYJXIb++1ERgAsIJRkhEtSYBLAUDkstzKis/pbFOstAJDkAIAPene3Py8mxS7rQE0AiTVGbe9FzYwHgDo0rG8LGcHt2fdvmAAYgHWnFgLSFtKgNsIgCKAec8LyUygQ28CCUQAwstyfvi/dkQqpTP5BHt/we8L0gBognk7snoXdchIwCjRYQCXhASCy3Ltbmo4huMcvl6gB602V4LKa60AftwwZkRmc7md1BEAIoI3DGA8AGiqAtCuY5nbFmDRbIA1xZbK6QHEoFjtSZPnBgCzan/CQgEmIwBNwW0/28plW2JxZHUefGl4SlczwE0mgTlb8Ehm6owSZWxGFEoAz4gOB+h2ItB27AlAi/gZEexsagegBC4aKxWqm0EtGoV9gfOkAmA8AJBFJzc1fNHtbIZ5/0EzNeNUEsB5yTqWmTMiBfD8uTfjjNehiaMBmiK3/XhftS+asdP2Ujm1BOywXu9Kk+d1Voe8J9IEDAAI6N/rjAAsd2+5mduLeNarAXIQj4MmltQAPqcDO1YqBCKYMeufGMWSzZtaAWhvuBnc7uodNiOix3S1AvzGhv9jONZWoFToeR0BsPlECDDOAMLLctCQuUUnz/s3PGAWDu9tTwPARs+rUKATCg3w3Jix8aQoAvrbWHPiAJ2xqpm54eGG/+tghuGYOyKVEaWXwG2WVQOAJng4w8MxAwCCowB6zCWbFZOVwhHRMGxs8vk8PWLRBDUD8Jnb1pXuAoAz44zf2dDfnpyOSiAovNrLcrA1s+eMpl5QpEe9HTUDfKCPeNA3MJJXz+uQwEZjBJgQEpgMJBCtmvn7EfqIqNG8JmRO2ysVsr1PB0DHnquDRgSwMnBmPCoBoN4xeaQK9cDzBbaz2gQDM/xf59XstP24ANaTJq+QgCSluLnkRjwdAnRGAKwduwedGlvo4360apYCICRQIgAAokSjNiMKVIgBYBNLWDVb3vKvCdnHeNATFfO+iaXjmAD399zTeLtGAmr5dSStDgC0DI4AsO8v2JROJ0SNaAU5FY55S2k6ABoL5k3BQwGgCGwsgJQuEwBMC4BI0YldE0IRuA0+dsXS2nGtAOE7KnjajgDWlZJDLvrbfYEEVsO6pbhf4HPS3LBP6SrHBzAvSNho5gCcGWcwKRUA6jscgHbn+yOWlqyRADSx0EOumgGqPIekAYgIfEZ0uAqtBoXXXqpDdl+QxZdg8Kx6IE970WoH+DV4TGjPSODVc20GdZQgM3qEBMKqmXyLx1iBPevVBPQdleMBuPcV55IzXgR1dnesdYgB9IUA8aoZGEF3Oz/kajQ3v10DwjEAWCxAAENQhzsbt6/hAH1ShSISYPcVTVadxYnPeEyXp++8HgcACMyrZgbgFU3pIgAxFQrKfr3yVTM//8C88zrgOyhqBYg+k7pHJABJdZ3fWgoJ9AUAqxEAS0D2Ba5olrNVs+ZUAL9WeZctaXj1ihB4Mz4cIFJ0olfNureoCODyfQ5G0w0MpAWo8hbPXvL2TMMrGgusCI4AuBSpW5L7iv6EZVGesJiT0poB7CMeN9lDrwqg4YyVQV0d2RfQ38bSa+EL6pbWjO0ro/ZVM33Ua7tiw6nhXwwgdEgTJA0EwAczEQegcln4grIfv3y/Yg6r4ajXetJiMS3ARuTJ7HklgbfeEfkTlgyTwFIIcGk1VrPxU0CMEg2ZlM7FgpMDQB1K3r61SqRFMOOU6CiAeNVMXRPyKd0WTJLJek9kr6jIqeFfAiAfCyYAVgROApCT7jAArUOHSmCnnz0sR0ZQmEYo7EvWOoSOaCcFwEbkgUgCwHbHM3S+UFSFmAR2/scvTvf4FyJtyUm35+dsuaDS/HdNBDufq7xafhsA3hJHZMPxzENe9vN3XieDOyryFZJuEo4X2THdcH5AvloeNrebq3LsXbYN8hYPE0HydYPXIX3EQg5KXen4TXBZ7lLVKSC9V9n2Xp0RYUcpdvXmeD6Bz66TdkD5EIx4UmuDPApmMiIFoL8zr0JHNMpKNr6hMXJTK3Lt2N++zw5lSeXVNEJVqjTFivcXnAjUc8cbggB06Pvkh7doBcaOZzhBJuO7eoPLcqvRiVC9JBbgGdGQbiSCR0a1DuVZUho8+h15l60eX7jciLyvOP9zcvEtfN6M3eZYFi59QyOfCBUSuFiw5c6qRTiuxCqvfKCSePPbv3DJEor/JhcQoAFE8JzqkG+mm4jc1BI3PDrD8QFudwyDTFr8E5f20XJ2TagcAaCPCd0KH5bTAOeTc+CGGpwnreN7S2EF4RWV6N119uA0tAMaR+SOesVrx2IUjjCCevqgE48FZ5PkB03QYD3pKy+CUWcEoQ6FExw6xSSZbvbEpTkptTo0wEpOUPzugotOZHD7yAh71uxW9IXL75MkuQgieIt2/JzvCzLmkOiNuGQzHYaCzsj0ADoFBBuhWvyFS1K9D68dSzvmj4JZR3T/pn6r/tzXigBCQYM8I/Jn1RM0mkFr+Lgc1ivvrnf3+Ld2dTg2g93srd2KLPuVutq6+INOVIn8w/HUju/DU/VGBN4K2M4mM5rx7YDRO6+R2/dugIM95HL9jHjp1T043UzfeS2bF7XoTKup8JlUEo7/C6+kaxEoGZwhjkgXv932PsODWeCIBqvMMcE3Oo0VqO78LNzwwO29FUFrs7yi0sVH04kntWhGpCRwFh+qv9jAHFEd2dk85AR0iMZ4tUkyfCgXefMbktJh91gwb4QqtZXL7qVaP1BJvg/J/NA3ifl+eIvhuOEMjWYPaTSeYH3J05PislzVKSDkxWlWeWV31+1wPRMLxsiL01PyvWZvxje/t+tPzn5tglkQjh9Cc/tocF9RXlcMCOzmuHur3T7XnDXPHZv+A1q9b+X3nIIntab4q+WoQ1aB9HehoUEQ0PaDCZFPVLu7HgvH/oBCl/1cb3swggK66crWjiMv1dbLh+N/PZ+Q7yK4UgVwRuSkojsftvfTPCn1OWln1YlQeM6IBDm8fS9naGBWXYpNFXsc2LFa/zdJIgioK3V7S1q4dKGgbzoY8hm8xdMrDygwHtuUiMw/cI6ozVkBHyUT1SGxfqVFX2tXeiYoFxg/5O85RUdQRMYR9Vo/5Mp+tvKaw51NkXaGi9bwMfFMqgvHhuC380nwgSWf8YfVM7yJxW7N+giAnwIyGLpSUXMy8w2h8NpiEqJguJ4xgi7yOqEPx/WaADYGG9+fTWKfUiPnSclpeyaMZtPs8n3VWGDD8TJvAWnM8XNGdnG6HBnKhXaAVqw96TdJle/sxbf2rLeO7syUK3WxoG+p2kSowWhG5Dc2K3rWLfSAtNh7TkFreLncVeaDWBDgsXuu+dnGN+eS6t/ZCz+8kmYAJkA6Mo0nnZyuFguqzDfEgoepF+ScCMKOTP62n8knzPb+50OXbxgu/vD1K5tUe0eUCcIxt4LB+Pa+2+V0hqCF7AuK3BGVydZsTD4WPPX9z9+cD1f//7PTgavzNdNIAAAAAElFTkSuQmCC
{{end}}

{{define "groupRow"}}
  <tr class="group">
  <td class="name">
    {{.GroupLabel}}: {{.Group}}
  </td>
  <td class="monthly-quantity"></td>
  <td class="unit"></td>
  <td class="price"></td>
  <td class="hourly-cost"></td>
  <td class="monthly-cost"></td>
  </tr>
{{end}}

{{define "resourceRows"}}
  <tr class="resource{{if eq .Indent 0}} top-level{{end}}">
    <td class="name">
      {{if gt .Indent 1}}{{repeat (int (add .Indent -1)) "&nbsp;&nbsp;&nbsp;&nbsp;" | safeHTML}}{{end}}
      {{if gt .Indent 0}}<span class="arrow">&#8627;</span>{{end}}
      {{.Resource.Name}}
    </td>
    <td class="monthly-quantity"></td>
    <td class="unit"></td>
    <td class="price"></td>
    <td class="hourly-cost">{{.Resource.HourlyCost | formatCost2DP}}</td>
    <td class="monthly-cost">{{.Resource.MonthlyCost | formatCost2DP}}</td>
  </tr>
  {{ if .Resource.Tags}}
    <tr class="tags">
      <td class="name">
        {{$tags := list}}
        {{range $k, $v := .Resource.Tags}}
          {{$t := list $k "=" $v | join "" }}
          {{$tags = append $tags $t}}
        {{end}}
        <span class="label">Tags:</span>
        <span>{{$tags | join ", "}}</span>
      </td>
      <td class="monthly-quantity"></td>
      <td class="unit"></td>
      <td class="price"></td>
      <td class="hourly-cost"></td>
      <td class="monthly-cost"></td>
    </tr>
  {{end}}
  {{$ident := add .Indent 1}}
  {{range .Resource.CostComponents}}
    {{template "costComponentRow" dict "CostComponent" . "Indent" $ident}}
  {{end}}
  {{range .Resource.SubResources}}
    {{template "resourceRows" dict "Resource" . "Indent" $ident}}
  {{end}}
{{end}}

{{define "costComponentRow"}}
  <tr class="cost-component">
    <td class="name">
      {{if gt .Indent 1}}{{repeat (int (add .Indent -1)) "&nbsp;&nbsp;&nbsp;&nbsp;" | safeHTML}}{{end}}
      {{if gt .Indent 0}}<span class="arrow">&#8627;</span>{{end}}
      {{.CostComponent.Name}}
    </td>
    <td class="monthly-quantity">{{.CostComponent.MonthlyQuantity | formatQuantity }}</td>
    <td class="unit">{{.CostComponent.Unit}}</td>
    <td class="price">{{.CostComponent.Price | formatPrice }}</td>
    <td class="hourly-cost">{{.CostComponent.HourlyCost | formatCost2DP}}</td>
    <td class="monthly-cost">{{.CostComponent.MonthlyCost | formatCost2DP}}</td>
  </tr>
{{end}}

<!doctype html>
<html>
  <head>
    <title>Infracost cost report</title>
    <style>
      {{template "style"}}
    </style>
    <link id="favicon" rel="shortcut icon" type="image/png" href="data:image/png;base64,{{template "faviconBase64"}}">
  </head>

  <body>
    <div class="metadata">
      <ul>
        <li>
          <span class="label">Generated by:</span>
          <span class="value"><a href="https://infracost.io" target="_blank">Infracost</a></span>
        </li>
        <li>
          <span class="label">Time generated:</span>
          <span class="value">{{.Root.TimeGenerated | date "2006-01-02 15:04:05 MST"}}</span>
        </li>
      </ul>
    </div>

    <table>
      <thead>
        <th class="name">Name</th>
        <th class="monthly-quantity">Monthly quantity</th>
        <th class="unit">Unit</th>
        <th class="price">Price</th>
        <th class="hourly-cost">Hourly cost</th>
        <th class="monthly-cost">Monthly cost</th>
      </thead>
      <tbody>
        {{$groupLabel := .Options.GroupLabel}}
        {{$groupKey := .Options.GroupKey}}
        {{$prevGroup := ""}}
        {{range .Root.Resources}}
          {{$group := index .Metadata $groupKey}}
          {{if ne $group $prevGroup}}
            {{template "groupRow" dict "GroupLabel" $groupLabel "Group" $group}}
          {{end}}
          {{template "resourceRows" dict "Resource" . "Indent" 0}}
          {{$prevGroup = $group}}
        {{end}}
        <tr class="spacer"><td colspan="6"></td></tr>
        <tr class="total">
          <td class="name">Overall total</td>
          <td class="monthly-quantity"></td>
          <td class="unit"></td>
          <td class="price"></td>
          <td class="hourly-cost">{{.Root.TotalHourlyCost | formatCost2DP}}</td>
          <td class="monthly-cost">{{.Root.TotalMonthlyCost | formatCost2DP}}</td>
        </tr>
      </tbody>
    </table>

    <div class="warnings">
      <p>{{.UnsupportedResourcesMessage | replaceNewLines}}</p>
    </div>
  </body>
</html>`
